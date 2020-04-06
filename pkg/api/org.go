package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/metrics"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

// GET /api/org
func GetOrgCurrent(c *m.ReqContext) Response {
	return getOrgHelper(c.OrgId)
}

// GET /api/orgs/:orgId
func GetOrgByID(c *m.ReqContext) Response {
	return getOrgHelper(c.ParamsInt64(":orgId"))
}

// Get /api/orgs/name/:name
func GetOrgByName(c *m.ReqContext) Response {
	query := m.GetOrgByNameQuery{Name: c.Params(":name")}
	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrOrgNotFound {
			return Error(404, "找不到组织", err)
		}

		return Error(500, "无法获取组织", err)
	}
	org := query.Result
	result := m.OrgDetailsDTO{
		Id:   org.Id,
		Name: org.Name,
		Address: m.Address{
			Address1: org.Address1,
			Address2: org.Address2,
			City:     org.City,
			ZipCode:  org.ZipCode,
			State:    org.State,
			Country:  org.Country,
		},
	}

	return JSON(200, &result)
}

func getOrgHelper(orgID int64) Response {
	query := m.GetOrgByIdQuery{Id: orgID}

	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrOrgNotFound {
			return Error(404, "找不到组织", err)
		}

		return Error(500, "无法获取组织", err)
	}

	org := query.Result
	result := m.OrgDetailsDTO{
		Id:   org.Id,
		Name: org.Name,
		Address: m.Address{
			Address1: org.Address1,
			Address2: org.Address2,
			City:     org.City,
			ZipCode:  org.ZipCode,
			State:    org.State,
			Country:  org.Country,
		},
	}

	return JSON(200, &result)
}

// POST /api/orgs
func CreateOrg(c *m.ReqContext, cmd m.CreateOrgCommand) Response {
	if !c.IsSignedIn || (!setting.AllowUserOrgCreate && !c.IsGrafanaAdmin) {
		return Error(403, "拒绝访问", nil)
	}

	cmd.UserId = c.UserId
	if err := bus.Dispatch(&cmd); err != nil {
		if err == m.ErrOrgNameTaken {
			return Error(409, "组织名称已被使用", err)
		}
		return Error(500, "创建组织失败", err)
	}

	metrics.MApiOrgCreate.Inc()

	return JSON(200, &util.DynMap{
		"orgId":   cmd.Result.Id,
		"message": "组织创建完成",
	})
}

// PUT /api/org
func UpdateOrgCurrent(c *m.ReqContext, form dtos.UpdateOrgForm) Response {
	return updateOrgHelper(form, c.OrgId)
}

// PUT /api/orgs/:orgId
func UpdateOrg(c *m.ReqContext, form dtos.UpdateOrgForm) Response {
	return updateOrgHelper(form, c.ParamsInt64(":orgId"))
}

func updateOrgHelper(form dtos.UpdateOrgForm, orgID int64) Response {
	cmd := m.UpdateOrgCommand{Name: form.Name, OrgId: orgID}
	if err := bus.Dispatch(&cmd); err != nil {
		if err == m.ErrOrgNameTaken {
			return Error(400, "组织名称已被使用", err)
		}
		return Error(500, "更新组织失败", err)
	}

	return Success("组织更新完成")
}

// PUT /api/org/address
func UpdateOrgAddressCurrent(c *m.ReqContext, form dtos.UpdateOrgAddressForm) Response {
	return updateOrgAddressHelper(form, c.OrgId)
}

// PUT /api/orgs/:orgId/address
func UpdateOrgAddress(c *m.ReqContext, form dtos.UpdateOrgAddressForm) Response {
	return updateOrgAddressHelper(form, c.ParamsInt64(":orgId"))
}

func updateOrgAddressHelper(form dtos.UpdateOrgAddressForm, orgID int64) Response {
	cmd := m.UpdateOrgAddressCommand{
		OrgId: orgID,
		Address: m.Address{
			Address1: form.Address1,
			Address2: form.Address2,
			City:     form.City,
			State:    form.State,
			ZipCode:  form.ZipCode,
			Country:  form.Country,
		},
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更新组织地址", err)
	}

	return Success("地址更新完成")
}

// GET /api/orgs/:orgId
func DeleteOrgByID(c *m.ReqContext) Response {
	if err := bus.Dispatch(&m.DeleteOrgCommand{Id: c.ParamsInt64(":orgId")}); err != nil {
		if err == m.ErrOrgNotFound {
			return Error(404, "无法删除组织。 找不到ID", nil)
		}
		return Error(500, "无法更新组织", err)
	}
	return Success("组织删除成功")
}

func SearchOrgs(c *m.ReqContext) Response {
	query := m.SearchOrgsQuery{
		Query: c.Query("query"),
		Name:  c.Query("name"),
		Page:  0,
		Limit: 1000,
	}

	if err := bus.Dispatch(&query); err != nil {
		return Error(500, "无法搜索到组织", err)
	}

	return JSON(200, query.Result)
}
