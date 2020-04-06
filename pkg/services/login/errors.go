package login

import "errors"

var (
	ErrInvalidCredentials = errors.New("账号或密码错误")
	ErrUsersQuotaReached  = errors.New("用户配额达到")
	ErrGettingUserQuota   = errors.New("获取用户配额时出错")
)
