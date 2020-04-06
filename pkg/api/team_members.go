package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/teamguardian"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

// GET /api/teams/:teamId/members
func GetTeamMembers(c *m.ReqContext) Response {
	query := m.GetTeamMembersQuery{OrgId: c.OrgId, TeamId: c.ParamsInt64(":teamId")}

	if err := bus.Dispatch(&query); err != nil {
		return Error(500, "无法获得组织成员", err)
	}

	for _, member := range query.Result {
		member.AvatarUrl = dtos.GetGravatarUrl(member.Email)
		member.Labels = []string{}

		if setting.IsEnterprise && member.External {
			authProvider := GetAuthProviderLabel(member.AuthModule)
			member.Labels = append(member.Labels, authProvider)
		}
	}

	return JSON(200, query.Result)
}

// POST /api/teams/:teamId/members
func (hs *HTTPServer) AddTeamMember(c *m.ReqContext, cmd m.AddTeamMemberCommand) Response {
	cmd.OrgId = c.OrgId
	cmd.TeamId = c.ParamsInt64(":teamId")

	if err := teamguardian.CanAdmin(hs.Bus, cmd.OrgId, cmd.TeamId, c.SignedInUser); err != nil {
		return Error(403, "不允许添加组织成员", err)
	}

	if err := hs.Bus.Dispatch(&cmd); err != nil {
		if err == m.ErrTeamNotFound {
			return Error(404, "组织未找到", nil)
		}

		if err == m.ErrTeamMemberAlreadyAdded {
			return Error(400, "用户已添加到此组织中", nil)
		}

		return Error(500, "无法将成员添加到组织", err)
	}

	return JSON(200, &util.DynMap{
		"message": "会员加入了组织",
	})
}

// PUT /:teamId/members/:userId
func (hs *HTTPServer) UpdateTeamMember(c *m.ReqContext, cmd m.UpdateTeamMemberCommand) Response {
	teamId := c.ParamsInt64(":teamId")
	orgId := c.OrgId

	if err := teamguardian.CanAdmin(hs.Bus, orgId, teamId, c.SignedInUser); err != nil {
		return Error(403, "不允许更新组织成员", err)
	}

	if c.OrgRole != m.ROLE_ADMIN {
		cmd.ProtectLastAdmin = true
	}

	cmd.TeamId = teamId
	cmd.UserId = c.ParamsInt64(":userId")
	cmd.OrgId = orgId

	if err := hs.Bus.Dispatch(&cmd); err != nil {
		if err == m.ErrTeamMemberNotFound {
			return Error(404, "未找到组织成员。", nil)
		}
		return Error(500, "无法更新组织成员。", err)
	}
	return Success("组织成员已更新")
}

// DELETE /api/teams/:teamId/members/:userId
func (hs *HTTPServer) RemoveTeamMember(c *m.ReqContext) Response {
	orgId := c.OrgId
	teamId := c.ParamsInt64(":teamId")
	userId := c.ParamsInt64(":userId")

	if err := teamguardian.CanAdmin(hs.Bus, orgId, teamId, c.SignedInUser); err != nil {
		return Error(403, "不允许删除组织成员", err)
	}

	protectLastAdmin := false
	if c.OrgRole != m.ROLE_ADMIN {
		protectLastAdmin = true
	}

	if err := hs.Bus.Dispatch(&m.RemoveTeamMemberCommand{OrgId: orgId, TeamId: teamId, UserId: userId, ProtectLastAdmin: protectLastAdmin}); err != nil {
		if err == m.ErrTeamNotFound {
			return Error(404, "组织未找到", nil)
		}

		if err == m.ErrTeamMemberNotFound {
			return Error(404, "未找到组织成员", nil)
		}

		return Error(500, "无法从组织中删除成员", err)
	}
	return Success("组织成员已删除")
}
