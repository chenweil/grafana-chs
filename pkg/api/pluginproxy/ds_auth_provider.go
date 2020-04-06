package pluginproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
	"github.com/grafana/grafana/pkg/util"
	"golang.org/x/oauth2/google"
)

//ApplyRoute should use the plugin route data to set auth headers and custom headers
func ApplyRoute(ctx context.Context, req *http.Request, proxyPath string, route *plugins.AppPluginRoute, ds *m.DataSource) {
	proxyPath = strings.TrimPrefix(proxyPath, route.Path)

	data := templateData{
		JsonData:       ds.JsonData.Interface().(map[string]interface{}),
		SecureJsonData: ds.SecureJsonData.Decrypt(),
	}

	interpolatedURL, err := InterpolateString(route.Url, data)
	if err != nil {
		logger.Error("插入代理地址时出错", "error", err)
		return
	}

	routeURL, err := url.Parse(interpolatedURL)
	if err != nil {
		logger.Error("解析插件路由地址时出错", "error", err)
		return
	}

	req.URL.Scheme = routeURL.Scheme
	req.URL.Host = routeURL.Host
	req.Host = routeURL.Host
	req.URL.Path = util.JoinURLFragments(routeURL.Path, proxyPath)

	if err := addHeaders(&req.Header, route, data); err != nil {
		logger.Error("无法呈现插件头信息", "error", err)
	}

	tokenProvider := newAccessTokenProvider(ds, route)

	if route.TokenAuth != nil {
		if token, err := tokenProvider.getAccessToken(data); err != nil {
			logger.Error("无法获得访问令牌", "error", err)
		} else {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	authenticationType := ds.JsonData.Get("authenticationType").MustString("jwt")
	if route.JwtTokenAuth != nil && authenticationType == "jwt" {
		if token, err := tokenProvider.getJwtAccessToken(ctx, data); err != nil {
			logger.Error("无法获得访问令牌", "error", err)
		} else {
			req.Header.Set("授权", fmt.Sprintf("Bearer %s", token))
		}
	}

	if authenticationType == "gce" {
		tokenSrc, err := google.DefaultTokenSource(ctx, route.JwtTokenAuth.Scopes...)
		if err != nil {
			logger.Error("无法从元数据服务器获取默认令牌", "error", err)
		} else {
			token, err := tokenSrc.Token()
			if err != nil {
				logger.Error("法从元数据服务器的默认访问令牌", "error", err)
			} else {
				req.Header.Set("授权", fmt.Sprintf("Bearer %s", token.AccessToken))
			}
		}
	}

	logger.Info("Requesting", "url", req.URL.String())
}

func addHeaders(reqHeaders *http.Header, route *plugins.AppPluginRoute, data templateData) error {
	for _, header := range route.Headers {
		interpolated, err := InterpolateString(header.Content, data)
		if err != nil {
			return err
		}
		reqHeaders.Add(header.Name, interpolated)
	}

	return nil
}
