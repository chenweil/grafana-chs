package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/rendering"
	"github.com/grafana/grafana/pkg/util"
)

func (hs *HTTPServer) RenderToPng(c *m.ReqContext) {
	queryReader, err := util.NewURLQueryReader(c.Req.URL)
	if err != nil {
		c.Handle(400, "渲染参数错误", err)
		return
	}

	queryParams := fmt.Sprintf("?%s", c.Req.URL.RawQuery)

	width, err := strconv.Atoi(queryReader.Get("width", "800"))
	if err != nil {
		c.Handle(400, "Render parameters error", fmt.Errorf("无法将宽度解析为int: %s", err))
		return
	}

	height, err := strconv.Atoi(queryReader.Get("height", "400"))
	if err != nil {
		c.Handle(400, "Render parameters error", fmt.Errorf("无法将高度解析为int: %s", err))
		return
	}

	timeout, err := strconv.Atoi(queryReader.Get("timeout", "60"))
	if err != nil {
		c.Handle(400, "Render parameters error", fmt.Errorf("无法将超时解析为int: %s", err))
		return
	}

	result, err := hs.RenderService.Render(c.Req.Context(), rendering.Opts{
		Width:           width,
		Height:          height,
		Timeout:         time.Duration(timeout) * time.Second,
		OrgId:           c.OrgId,
		UserId:          c.UserId,
		OrgRole:         c.OrgRole,
		Path:            c.Params("*") + queryParams,
		Timezone:        queryReader.Get("tz", ""),
		Encoding:        queryReader.Get("encoding", ""),
		ConcurrentLimit: 30,
	})

	if err != nil && err == rendering.ErrTimeout {
		c.Handle(500, err.Error(), err)
		return
	}

	if err != nil && err == rendering.ErrPhantomJSNotInstalled {
		if strings.HasPrefix(runtime.GOARCH, "arm") {
			c.Handle(500, "渲染失败 - 默认情况下，PhantomJS不包含在arm构建中", err)
		} else {
			c.Handle(500, "渲染失败 - 未正确安装PhantomJS", err)
		}
		return
	}

	if err != nil {
		c.Handle(500, "渲染失败.", err)
		return
	}

	c.Resp.Header().Set("Content-Type", "image/png")
	http.ServeFile(c.Resp, c.Req.Request, result.FilePath)
}
