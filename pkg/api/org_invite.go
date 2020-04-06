package api

import (
	"fmt"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/events"
	"github.com/grafana/grafana/pkg/infra/metrics"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

func GetPendingOrgInvites(c *m.ReqContext) Response {
	query := m.GetTempUsersQuery{OrgId: c.OrgId, Status: m.TmpUserInvitePending}

	if err := bus.Dispatch(&query); err != nil {
		return Error(500, "无法从数据库中获得邀请", err)
	}

	for _, invite := range query.Result {
		invite.Url = setting.ToAbsUrl("invite/" + invite.Code)
	}

	return JSON(200, query.Result)
}

func AddOrgInvite(c *m.ReqContext, inviteDto dtos.AddInviteForm) Response {
	if setting.DisableLoginForm {
		return Error(400, "禁用登录时无法邀请。", nil)
	}

	if !inviteDto.Role.IsValid() {
		return Error(400, "指定的角色无效", nil)
	}

	// first try get existing user
	userQuery := m.GetUserByLoginQuery{LoginOrEmail: inviteDto.LoginOrEmail}
	if err := bus.Dispatch(&userQuery); err != nil {
		if err != m.ErrUserNotFound {
			return Error(500, "无法查询数据库以进行现有用户检查", err)
		}
	} else {
		return inviteExistingUserToOrg(c, userQuery.Result, &inviteDto)
	}

	cmd := m.CreateTempUserCommand{}
	cmd.OrgId = c.OrgId
	cmd.Email = inviteDto.LoginOrEmail
	cmd.Name = inviteDto.Name
	cmd.Status = m.TmpUserInvitePending
	cmd.InvitedByUserId = c.UserId
	cmd.Code = util.GetRandomString(30)
	cmd.Role = inviteDto.Role
	cmd.RemoteAddr = c.Req.RemoteAddr

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法将邀请保存到数据库", err)
	}

	// send invite email
	if inviteDto.SendEmail && util.IsEmail(inviteDto.LoginOrEmail) {
		emailCmd := m.SendEmailCommand{
			To:       []string{inviteDto.LoginOrEmail},
			Template: "new_user_invite.html",
			Data: map[string]interface{}{
				"Name":      util.StringsFallback2(cmd.Name, cmd.Email),
				"OrgName":   c.OrgName,
				"Email":     c.Email,
				"LinkUrl":   setting.ToAbsUrl("invite/" + cmd.Code),
				"InvitedBy": util.StringsFallback3(c.Name, c.Email, c.Login),
			},
		}

		if err := bus.Dispatch(&emailCmd); err != nil {
			if err == m.ErrSmtpNotEnabled {
				return Error(412, err.Error(), err)
			}
			return Error(500, "无法发送电子邮件邀请", err)
		}

		emailSentCmd := m.UpdateTempUserWithEmailSentCommand{Code: cmd.Result.Code}
		if err := bus.Dispatch(&emailSentCmd); err != nil {
			return Error(500, "无法使用电子邮件发送信息更新邀请", err)
		}

		return Success(fmt.Sprintf("已发送邀请到 %s", inviteDto.LoginOrEmail))
	}

	return Success(fmt.Sprintf("创建邀请 %s", inviteDto.LoginOrEmail))
}

func inviteExistingUserToOrg(c *m.ReqContext, user *m.User, inviteDto *dtos.AddInviteForm) Response {
	// user exists, add org role
	createOrgUserCmd := m.AddOrgUserCommand{OrgId: c.OrgId, UserId: user.Id, Role: inviteDto.Role}
	if err := bus.Dispatch(&createOrgUserCmd); err != nil {
		if err == m.ErrOrgUserAlreadyAdded {
			return Error(412, fmt.Sprintf("用户 %s 已添加到组织中", inviteDto.LoginOrEmail), err)
		}
		return Error(500, "尝试创建组织用户时出错", err)
	}

	if inviteDto.SendEmail && util.IsEmail(user.Email) {
		emailCmd := m.SendEmailCommand{
			To:       []string{user.Email},
			Template: "invited_to_org.html",
			Data: map[string]interface{}{
				"Name":      user.NameOrFallback(),
				"OrgName":   c.OrgName,
				"InvitedBy": util.StringsFallback3(c.Name, c.Email, c.Login),
			},
		}

		if err := bus.Dispatch(&emailCmd); err != nil {
			return Error(500, "无法发送电子邮件invite_to_org", err)
		}
	}

	return Success(fmt.Sprintf("现有的用户 %s 已添加到组织 %s", user.NameOrFallback(), c.OrgName))
}

func RevokeInvite(c *m.ReqContext) Response {
	if ok, rsp := updateTempUserStatus(c.Params(":code"), m.TmpUserRevoked); !ok {
		return rsp
	}

	return Success("邀请已撤销")
}

func GetInviteInfoByCode(c *m.ReqContext) Response {
	query := m.GetTempUserByCodeQuery{Code: c.Params(":code")}

	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrTempUserNotFound {
			return Error(404, "邀请未找到", nil)
		}
		return Error(500, "无法获得邀请", err)
	}

	invite := query.Result

	return JSON(200, dtos.InviteInfo{
		Email:     invite.Email,
		Name:      invite.Name,
		Username:  invite.Email,
		InvitedBy: util.StringsFallback3(invite.InvitedByName, invite.InvitedByLogin, invite.InvitedByEmail),
	})
}

func (hs *HTTPServer) CompleteInvite(c *m.ReqContext, completeInvite dtos.CompleteInviteForm) Response {
	query := m.GetTempUserByCodeQuery{Code: completeInvite.InviteCode}

	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrTempUserNotFound {
			return Error(404, "邀请未找到d", nil)
		}
		return Error(500, "无法获得邀请", err)
	}

	invite := query.Result
	if invite.Status != m.TmpUserInvitePending {
		return Error(412, fmt.Sprintf("邀请不能在 %s 状态中使用", invite.Status), nil)
	}

	cmd := m.CreateUserCommand{
		Email:        completeInvite.Email,
		Name:         completeInvite.Name,
		Login:        completeInvite.Username,
		Password:     completeInvite.Password,
		SkipOrgSetup: true,
	}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法创建用户", err)
	}

	user := &cmd.Result

	bus.Publish(&events.SignUpCompleted{
		Name:  user.NameOrFallback(),
		Email: user.Email,
	})

	if ok, rsp := applyUserInvite(user, invite, true); !ok {
		return rsp
	}

	hs.loginUserWithUser(user, c)

	metrics.MApiUserSignUpCompleted.Inc()
	metrics.MApiUserSignUpInvite.Inc()

	return Success("用户已创建并登录")
}

func updateTempUserStatus(code string, status m.TempUserStatus) (bool, Response) {
	// update temp user status
	updateTmpUserCmd := m.UpdateTempUserStatusCommand{Code: code, Status: status}
	if err := bus.Dispatch(&updateTmpUserCmd); err != nil {
		return false, Error(500, "无法更新邀请状态", err)
	}

	return true, nil
}

func applyUserInvite(user *m.User, invite *m.TempUserDTO, setActive bool) (bool, Response) {
	// add to org
	addOrgUserCmd := m.AddOrgUserCommand{OrgId: invite.OrgId, UserId: user.Id, Role: invite.Role}
	if err := bus.Dispatch(&addOrgUserCmd); err != nil {
		if err != m.ErrOrgUserAlreadyAdded {
			return false, Error(500, "尝试创建组织用户时出错", err)
		}
	}

	// update temp user status
	if ok, rsp := updateTempUserStatus(invite.Code, m.TmpUserCompleted); !ok {
		return false, rsp
	}

	if setActive {
		// set org to active
		if err := bus.Dispatch(&m.SetUsingOrgCommand{OrgId: invite.OrgId, UserId: user.Id}); err != nil {
			return false, Error(500, "无法将组织设置为活动状态", err)
		}
	}

	return true, nil
}
