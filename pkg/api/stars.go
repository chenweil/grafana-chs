package api

import (
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
)

func StarDashboard(c *m.ReqContext) Response {
	if !c.IsSignedIn {
		return Error(412, "您需要登录星标表", nil)
	}

	cmd := m.StarDashboardCommand{UserId: c.UserId, DashboardId: c.ParamsInt64(":id")}

	if cmd.DashboardId <= 0 {
		return Error(400, "缺少仪表板ID", nil)
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法启动仪表板", err)
	}

	return Success("仪表板加入星标!")
}

func UnstarDashboard(c *m.ReqContext) Response {

	cmd := m.UnstarDashboardCommand{UserId: c.UserId, DashboardId: c.ParamsInt64(":id")}

	if cmd.DashboardId <= 0 {
		return Error(400, "缺少仪表板ID", nil)
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法取消选中仪表板", err)
	}

	return Success("仪表板取消星标")
}
