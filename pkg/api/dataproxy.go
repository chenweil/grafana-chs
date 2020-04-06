package api

import (
	"github.com/grafana/grafana/pkg/api/pluginproxy"
	"github.com/grafana/grafana/pkg/infra/metrics"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/plugins"
)

func (hs *HTTPServer) ProxyDataSourceRequest(c *m.ReqContext) {
	c.TimeRequest(metrics.MDataSourceProxyReqTimer)

	dsId := c.ParamsInt64(":id")
	ds, err := hs.DatasourceCache.GetDatasource(dsId, c.SignedInUser, c.SkipCache)
	if err != nil {
		if err == m.ErrDataSourceAccessDenied {
			c.JsonApiErr(403, "拒绝访问数据源", err)
			return
		}
		c.JsonApiErr(500, "无法加载数据源元数据", err)
		return
	}

	// find plugin
	plugin, ok := plugins.DataSources[ds.Type]
	if !ok {
		c.JsonApiErr(500, "无法找到数据源插件", err)
		return
	}

	// macaron does not include trailing slashes when resolving a wildcard path
	proxyPath := ensureProxyPathTrailingSlash(c.Req.URL.Path, c.Params("*"))

	proxy := pluginproxy.NewDataSourceProxy(ds, plugin, c, proxyPath, hs.Cfg)
	proxy.HandleRequest()
}

// ensureProxyPathTrailingSlash Check for a trailing slash in original path and makes
// sure that a trailing slash is added to proxy path, if not already exists.
func ensureProxyPathTrailingSlash(originalPath, proxyPath string) string {
	if len(proxyPath) > 1 {
		if originalPath[len(originalPath)-1] == '/' && proxyPath[len(proxyPath)-1] != '/' {
			return proxyPath + "/"
		}
	}

	return proxyPath
}
