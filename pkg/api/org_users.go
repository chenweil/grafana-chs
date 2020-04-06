package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
)

// POST /api/org/users
func AddOrgUserToCurrentOrg(c *m.ReqContext, cmd m.AddOrgUserCommand) Response {
	cmd.OrgId = c.OrgId
	return addOrgUserHelper(cmd)
}

// POST /api/orgs/:orgId/users
func AddOrgUser(c *m.ReqContext, cmd m.AddOrgUserCommand) Response {
	cmd.OrgId = c.ParamsInt64(":orgId")
	return addOrgUserHelper(cmd)
}

func addOrgUserHelper(cmd m.AddOrgUserCommand) Response {
	if !cmd.Role.IsValid() {
		return Error(400, "指定的角色无效", nil)
	}

	userQuery := m.GetUserByLoginQuery{LoginOrEmail: cmd.LoginOrEmail}
	err := bus.Dispatch(&userQuery)
	if err != nil {
		return Error(404, "找不到用户", nil)
	}

	userToAdd := userQuery.Result

	cmd.UserId = userToAdd.Id

	if err := bus.Dispatch(&cmd); err != nil {
		if err == m.ErrOrgUserAlreadyAdded {
			return Error(409, "用户已经是该组织的成员", nil)
		}
		return Error(500, "无法将用户添加到组织", err)
	}

	return Success("用户已添加到组织中")
}

// GET /api/org/users
func GetOrgUsersForCurrentOrg(c *m.ReqContext) Response {
	return getOrgUsersHelper(c.OrgId, c.Query("query"), c.QueryInt("limit"))
}

// GET /api/orgs/:orgId/users
func GetOrgUsers(c *m.ReqContext) Response {
	return getOrgUsersHelper(c.ParamsInt64(":orgId"), "", 0)
}

func getOrgUsersHelper(orgID int64, query string, limit int) Response {
	q := m.GetOrgUsersQuery{
		OrgId: orgID,
		Query: query,
		Limit: limit,
	}

	if err := bus.Dispatch(&q); err != nil {
		return Error(500, "无法获得帐户用户", err)
	}

	for _, user := range q.Result {
		user.AvatarUrl = dtos.GetGravatarUrl(user.Email)
	}

	return JSON(200, q.Result)
}

// PATCH /api/org/users/:userId
func UpdateOrgUserForCurrentOrg(c *m.ReqContext, cmd m.UpdateOrgUserCommand) Response {
	cmd.OrgId = c.OrgId
	cmd.UserId = c.ParamsInt64(":userId")
	return updateOrgUserHelper(cmd)
}

// PATCH /api/orgs/:orgId/users/:userId
func UpdateOrgUser(c *m.ReqContext, cmd m.UpdateOrgUserCommand) Response {
	cmd.OrgId = c.ParamsInt64(":orgId")
	cmd.UserId = c.ParamsInt64(":userId")
	return updateOrgUserHelper(cmd)
}

func updateOrgUserHelper(cmd m.UpdateOrgUserCommand) Response {
	if !cmd.Role.IsValid() {
		return Error(400, "指定的角色无效", nil)
	}

	if err := bus.Dispatch(&cmd); err != nil {
		if err == m.ErrLastOrgAdmin {
			return Error(400, "无法更改角色，因此没有组织管理员", nil)
		}
		return Error(500, "更新组织用户失败", err)
	}

	return Success("组织用户已更新")
}

// DELETE /api/org/users/:userId
func RemoveOrgUserForCurrentOrg(c *m.ReqContext) Response {
	return removeOrgUserHelper(&m.RemoveOrgUserCommand{
		UserId:                   c.ParamsInt64(":userId"),
		OrgId:                    c.OrgId,
		ShouldDeleteOrphanedUser: true,
	})
}

// DELETE /api/orgs/:orgId/users/:userId
func RemoveOrgUser(c *m.ReqContext) Response {
	return removeOrgUserHelper(&m.RemoveOrgUserCommand{
		UserId: c.ParamsInt64(":userId"),
		OrgId:  c.ParamsInt64(":orgId"),
	})
}

func removeOrgUserHelper(cmd *m.RemoveOrgUserCommand) Response {
	if err := bus.Dispatch(cmd); err != nil {
		if err == m.ErrLastOrgAdmin {
			return Error(400, "无法删除上一个组织管理员", nil)
		}
		return Error(500, "无法从组织中删除用户", err)
	}

	if cmd.UserWasDeleted {
		return Success("用户已删除")
	}

	return Success("用户已从组织中删除")
}
