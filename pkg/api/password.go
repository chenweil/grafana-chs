package api

import (
	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

func SendResetPasswordEmail(c *m.ReqContext, form dtos.SendResetPasswordEmailForm) Response {
	if setting.LDAPEnabled || setting.AuthProxyEnabled {
		return Error(401, "启用LDAP或验证代理时，不允许重置密码", nil)
	}
	if setting.DisableLoginForm {
		return Error(401, "禁用登录表单时不允许重置密码", nil)
	}

	userQuery := m.GetUserByLoginQuery{LoginOrEmail: form.UserOrEmail}

	if err := bus.Dispatch(&userQuery); err != nil {
		c.Logger.Info("Requested password reset for user that was not found", "user", userQuery.LoginOrEmail)
		return Error(200, "邮件已发送", err)
	}

	emailCmd := m.SendResetPasswordEmailCommand{User: userQuery.Result}
	if err := bus.Dispatch(&emailCmd); err != nil {
		return Error(500, "无法发送电子邮件", err)
	}

	return Success("邮件发送成功")
}

func ResetPassword(c *m.ReqContext, form dtos.ResetUserPasswordForm) Response {
	query := m.ValidateResetPasswordCodeQuery{Code: form.Code}

	if err := bus.Dispatch(&query); err != nil {
		if err == m.ErrInvalidEmailCode {
			return Error(400, "重置密码代码无效或过期", nil)
		}
		return Error(500, "验证电子邮件代码的未知错误", err)
	}

	if form.NewPassword != form.ConfirmPassword {
		return Error(400, "密码不匹配", nil)
	}

	cmd := m.ChangeUserPasswordCommand{}
	cmd.UserId = query.Result.Id
	cmd.NewPassword = util.EncodePassword(form.NewPassword, query.Result.Salt)

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法更改用户密码", err)
	}

	return Success("用户密码更改成功")
}
