package notifiers

import (
	"os"
	"strings"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"

	"github.com/grafana/grafana/pkg/services/alerting"
	"github.com/grafana/grafana/pkg/setting"
)

func init() {
	alerting.RegisterNotifier(&alerting.NotifierPlugin{
		Type:        "email",
		Name:        "邮件",
		Description: "使用服务器配置的SMTP设置发送通知",
		Factory:     NewEmailNotifier,
		OptionsTemplate: `
      <h3 class="page-heading">邮件地址</h3>
      <div class="gf-form">
         <textarea rows="7" class="gf-form-input width-27" required ng-model="ctrl.model.settings.addresses"></textarea>
      </div>
      <div class="gf-form">
      <span>您可以使用“;”分隔输入多个电子邮件地址</span>
      </div>
    `,
	})
}

// EmailNotifier is responsible for sending
// alert notifications over email.
type EmailNotifier struct {
	NotifierBase
	Addresses []string
	log       log.Logger
}

// NewEmailNotifier is the constructor function
// for the EmailNotifier.
func NewEmailNotifier(model *models.AlertNotification) (alerting.Notifier, error) {
	addressesString := model.Settings.Get("addresses").MustString()

	if addressesString == "" {
		return nil, alerting.ValidationError{Reason: "无法在设置中找到地址"}
	}

	// split addresses with a few different ways
	addresses := strings.FieldsFunc(addressesString, func(r rune) bool {
		switch r {
		case ',', ';', '\n':
			return true
		}
		return false
	})

	return &EmailNotifier{
		NotifierBase: NewNotifierBase(model),
		Addresses:    addresses,
		log:          log.New("alerting.notifier.email"),
	}, nil
}

// Notify sends the alert notification.
func (en *EmailNotifier) Notify(evalContext *alerting.EvalContext) error {
	en.log.Info("发送警报通知", "addresses", en.Addresses)

	ruleURL, err := evalContext.GetRuleURL()
	if err != nil {
		en.log.Error("获取规则链接失败", "error", err)
		return err
	}

	error := ""
	if evalContext.Error != nil {
		error = evalContext.Error.Error()
	}

	cmd := &models.SendEmailCommandSync{
		SendEmailCommand: models.SendEmailCommand{
			Subject: evalContext.GetNotificationTitle(),
			Data: map[string]interface{}{
				"Title":         evalContext.GetNotificationTitle(),
				"State":         evalContext.Rule.State,
				"Name":          evalContext.Rule.Name,
				"StateModel":    evalContext.GetStateModel(),
				"Message":       evalContext.Rule.Message,
				"Error":         error,
				"RuleUrl":       ruleURL,
				"ImageLink":     "",
				"EmbeddedImage": "",
				"AlertPageUrl":  setting.AppUrl + "alerting",
				"EvalMatches":   evalContext.EvalMatches,
			},
			To:           en.Addresses,
			Template:     "alert_notification.html",
			EmbededFiles: []string{},
		},
	}

	if evalContext.ImagePublicURL != "" {
		cmd.Data["ImageLink"] = evalContext.ImagePublicURL
	} else {
		file, err := os.Stat(evalContext.ImageOnDiskPath)
		if err == nil {
			cmd.EmbededFiles = []string{evalContext.ImageOnDiskPath}
			cmd.Data["EmbeddedImage"] = file.Name()
		}
	}

	err = bus.DispatchCtx(evalContext.Ctx, cmd)

	if err != nil {
		en.log.Error("无法发送警报通知电子邮件", "error", err)
		return err
	}

	return nil
}
