package login

import (
	"errors"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/ldap"
)

var (
	ErrEmailNotAllowed       = errors.New("未满足所需的电子邮件域")
	ErrInvalidCredentials    = errors.New("账号或密码错误")
	ErrNoEmail               = errors.New("登录提供商未返回电子邮件地址")
	ErrProviderDeniedRequest = errors.New("登录提供商拒绝登录请求")
	ErrSignUpNotAllowed      = errors.New("此适配器不允许注册")
	ErrTooManyLoginAttempts  = errors.New("用户连续错误登录尝试次数过多。 登录用户暂时被阻止")
	ErrPasswordEmpty         = errors.New("没有提供密码")
	ErrUserDisabled          = errors.New("用户被禁用")
)

func Init() {
	bus.AddHandler("auth", AuthenticateUser)
}

// AuthenticateUser authenticates the user via username & password
func AuthenticateUser(query *models.LoginUserQuery) error {
	if err := validateLoginAttempts(query.Username); err != nil {
		return err
	}

	if err := validatePasswordSet(query.Password); err != nil {
		return err
	}

	err := loginUsingGrafanaDB(query)
	if err == nil || (err != models.ErrUserNotFound && err != ErrInvalidCredentials && err != ErrUserDisabled) {
		return err
	}

	ldapEnabled, ldapErr := loginUsingLDAP(query)
	if ldapEnabled {
		if ldapErr == nil || ldapErr != ldap.ErrInvalidCredentials {
			return ldapErr
		}

		if err != ErrUserDisabled || ldapErr != ldap.ErrInvalidCredentials {
			err = ldapErr
		}
	}

	if err == ErrInvalidCredentials || err == ldap.ErrInvalidCredentials {
		saveInvalidLoginAttempt(query)
		return ErrInvalidCredentials
	}

	if err == models.ErrUserNotFound {
		return ErrInvalidCredentials
	}

	return err
}

func validatePasswordSet(password string) error {
	if len(password) == 0 {
		return ErrPasswordEmpty
	}

	return nil
}
