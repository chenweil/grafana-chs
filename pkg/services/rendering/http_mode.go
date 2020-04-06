package rendering

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var netTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

var netClient = &http.Client{
	Transport: netTransport,
}

func (rs *RenderingService) renderViaHttp(ctx context.Context, opts Opts) (*RenderResult, error) {
	filePath := rs.getFilePathForNewImage()

	rendererUrl, err := url.Parse(rs.Cfg.RendererUrl)
	if err != nil {
		return nil, err
	}

	queryParams := rendererUrl.Query()
	queryParams.Add("url", rs.getURL(opts.Path))
	queryParams.Add("renderKey", rs.getRenderKey(opts.OrgId, opts.UserId, opts.OrgRole))
	queryParams.Add("width", strconv.Itoa(opts.Width))
	queryParams.Add("height", strconv.Itoa(opts.Height))
	queryParams.Add("domain", rs.domain)
	queryParams.Add("timezone", isoTimeOffsetToPosixTz(opts.Timezone))
	queryParams.Add("encoding", opts.Encoding)
	queryParams.Add("timeout", strconv.Itoa(int(opts.Timeout.Seconds())))
	rendererUrl.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", rendererUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	reqContext, cancel := context.WithTimeout(ctx, opts.Timeout+time.Second*2)
	defer cancel()

	req = req.WithContext(reqContext)

	// make request to renderer server
	resp, err := netClient.Do(req)
	if err != nil {
		rs.log.Error("无法向远程呈现服务发送请求。", "error", err)
		return nil, fmt.Errorf("无法向远程呈现服务发送请求。 %s", err)
	}

	// save response to file
	defer resp.Body.Close()

	// check for timeout first
	if reqContext.Err() == context.DeadlineExceeded {
		rs.log.Info("渲染超时")
		return nil, ErrTimeout
	}

	// if we didn't get a 200 response, something went wrong.
	if resp.StatusCode != http.StatusOK {
		rs.log.Error("远程呈现请求失败", "error", resp.Status)
		return nil, fmt.Errorf("远程呈现请求失败。 %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// check that we didn't timeout while receiving the response.
		if reqContext.Err() == context.DeadlineExceeded {
			rs.log.Info("渲染超时")
			return nil, ErrTimeout
		}
		rs.log.Error("远程呈现请求失败", "error", err)
		return nil, fmt.Errorf("远程呈现请求失败。  %s", err)
	}

	return &RenderResult{FilePath: filePath}, err
}
