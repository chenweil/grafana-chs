package api

import (
	"github.com/grafana/grafana/pkg/services/ldap"
)

func (server *HTTPServer) ReloadLDAPCfg() Response {
	if !ldap.IsEnabled() {
		return Error(400, "LDAP未启用", nil)
	}

	err := ldap.ReloadConfig()
	if err != nil {
		return Error(500, "无法重新加载ldap配置。", err)
	}
	return Success("LDAP配置已重新加载")
}
