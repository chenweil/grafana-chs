package pluginproxy

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/grafana/grafana/pkg/setting"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/util"
)

type templateData struct {
	JsonData       map[string]interface{}
	SecureJsonData map[string]string
}

func getHeaders(route *plugins.AppPluginRoute, orgId int64, appID string) (http.Header, error) {
	result := http.Header{}

	query := m.GetPluginSettingByIdQuery{OrgId: orgId, PluginId: appID}

	if err := bus.Dispatch(&query); err != nil {
		return nil, err
	}

	data := templateData{
		JsonData:       query.Result.JsonData,
		SecureJsonData: query.Result.SecureJsonData.Decrypt(),
	}

	err := addHeaders(&result, route, data)
	return result, err
}

func updateURL(route *plugins.AppPluginRoute, orgId int64, appID string) (string, error) {
	query := m.GetPluginSettingByIdQuery{OrgId: orgId, PluginId: appID}
	if err := bus.Dispatch(&query); err != nil {
		return "", err
	}

	data := templateData{
		JsonData:       query.Result.JsonData,
		SecureJsonData: query.Result.SecureJsonData.Decrypt(),
	}
	interpolated, err := InterpolateString(route.Url, data)
	if err != nil {
		return "", err
	}
	return interpolated, err
}

// NewApiPluginProxy create a plugin proxy
func NewApiPluginProxy(ctx *m.ReqContext, proxyPath string, route *plugins.AppPluginRoute, appID string, cfg *setting.Cfg) *httputil.ReverseProxy {
	targetURL, _ := url.Parse(route.Url)

	director := func(req *http.Request) {

		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host

		req.URL.Path = util.JoinURLFragments(targetURL.Path, proxyPath)
		// clear cookie headers
		req.Header.Del("Cookie")
		req.Header.Del("Set-Cookie")

		// clear X-Forwarded Host/Port/Proto headers
		req.Header.Del("X-Forwarded-Host")
		req.Header.Del("X-Forwarded-Port")
		req.Header.Del("X-Forwarded-Proto")

		// set X-Forwarded-For header
		if req.RemoteAddr != "" {
			remoteAddr, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				remoteAddr = req.RemoteAddr
			}
			if req.Header.Get("X-Forwarded-For") != "" {
				req.Header.Set("X-Forwarded-For", req.Header.Get("X-Forwarded-For")+", "+remoteAddr)
			} else {
				req.Header.Set("X-Forwarded-For", remoteAddr)
			}
		}

		// Create a HTTP header with the context in it.
		ctxJSON, err := json.Marshal(ctx.SignedInUser)
		if err != nil {
			ctx.JsonApiErr(500, "未能为json编组上下文。", err)
			return
		}

		req.Header.Add("X-Grafana-Context", string(ctxJSON))

		if cfg.SendUserHeader && !ctx.SignedInUser.IsAnonymous {
			req.Header.Add("X-Grafana-User", ctx.SignedInUser.Login)
		}

		if len(route.Headers) > 0 {
			headers, err := getHeaders(route, ctx.OrgId, appID)
			if err != nil {
				ctx.JsonApiErr(500, "无法生成插件路由报头", err)
				return
			}

			for key, value := range headers {
				log.Trace("setting key %v value <redacted>", key)
				req.Header.Set(key, value[0])
			}
		}

		if len(route.Url) > 0 {
			interpolatedURL, err := updateURL(route, ctx.OrgId, appID)
			if err != nil {
				ctx.JsonApiErr(500, "无法插入插件路由网址", err)
			}
			targetURL, err := url.Parse(interpolatedURL)
			if err != nil {
				ctx.JsonApiErr(500, "无法解析自定义网址: %v", err)
				return
			}
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.Host = targetURL.Host
			req.URL.Path = util.JoinURLFragments(targetURL.Path, proxyPath)
		}

		// reqBytes, _ := httputil.DumpRequestOut(req, true);
		// log.Trace("代理插件请求: %s", string(reqBytes))
	}

	return &httputil.ReverseProxy{Director: director}
}
