package models

import "errors"

var ErrInvalidEmailCode = errors.New("电子邮件代码无效或过期")
var ErrSmtpNotEnabled = errors.New("SMTP未配置，请检查您的grafana.ini配置文件的[smtp]部分")

type SendEmailCommand struct {
	To           []string
	Template     string
	Subject      string
	Data         map[string]interface{}
	Info         string
	EmbededFiles []string
}

type SendEmailCommandSync struct {
	SendEmailCommand
}

type SendWebhookSync struct {
	Url         string
	User        string
	Password    string
	Body        string
	HttpMethod  string
	HttpHeader  map[string]string
	ContentType string
}

type SendResetPasswordEmailCommand struct {
	User *User
}

type ValidateResetPasswordCodeQuery struct {
	Code   string
	Result *User
}
