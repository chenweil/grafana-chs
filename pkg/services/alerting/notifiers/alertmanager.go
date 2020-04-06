package notifiers

import (
	"context"
	"regexp"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/alerting"
)

func init() {
	alerting.RegisterNotifier(&alerting.NotifierPlugin{
		Type:        "prometheus-alertmanager",
		Name:        "警报管理控制器",
		Description: "发送警报到警报管理系统",
		Factory:     NewAlertmanagerNotifier,
		OptionsTemplate: `
      <h3 class="page-heading">警报管理器配置</h3>
      <div class="gf-form">
        <span class="gf-form-label width-10">地址</span>
        <input type="text" required class="gf-form-input max-width-26" ng-model="ctrl.model.settings.url" placeholder="http://localhost:9093"></input>
      </div>
    `,
	})
}

// NewAlertmanagerNotifier returns a new Alertmanager notifier
func NewAlertmanagerNotifier(model *models.AlertNotification) (alerting.Notifier, error) {
	url := model.Settings.Get("url").MustString()
	if url == "" {
		return nil, alerting.ValidationError{Reason: "在设置中找不到地址属性"}
	}

	return &AlertmanagerNotifier{
		NotifierBase: NewNotifierBase(model),
		URL:          url,
		log:          log.New("alerting.notifier.prometheus-alertmanager"),
	}, nil
}

// AlertmanagerNotifier sends alert notifications to the alert manager
type AlertmanagerNotifier struct {
	NotifierBase
	URL string
	log log.Logger
}

// ShouldNotify returns true if the notifiers should be used depending on state
func (am *AlertmanagerNotifier) ShouldNotify(ctx context.Context, evalContext *alerting.EvalContext, notificationState *models.AlertNotificationState) bool {
	am.log.Debug("Should notify", "ruleId", evalContext.Rule.ID, "state", evalContext.Rule.State, "previousState", evalContext.PrevAlertState)

	// Do not notify when we become OK for the first time.
	if (evalContext.PrevAlertState == models.AlertStatePending) && (evalContext.Rule.State == models.AlertStateOK) {
		return false
	}

	// Notify on Alerting -> OK to resolve before alertmanager timeout.
	if (evalContext.PrevAlertState == models.AlertStateAlerting) && (evalContext.Rule.State == models.AlertStateOK) {
		return true
	}

	return evalContext.Rule.State == models.AlertStateAlerting
}

func (am *AlertmanagerNotifier) createAlert(evalContext *alerting.EvalContext, match *alerting.EvalMatch, ruleURL string) *simplejson.Json {
	alertJSON := simplejson.New()
	alertJSON.Set("startsAt", evalContext.StartTime.UTC().Format(time.RFC3339))
	if evalContext.Rule.State == models.AlertStateOK {
		alertJSON.Set("endsAt", time.Now().UTC().Format(time.RFC3339))
	}
	alertJSON.Set("generatorURL", ruleURL)

	// Annotations (summary and description are very commonly used).
	alertJSON.SetPath([]string{"annotations", "summary"}, evalContext.Rule.Name)
	description := ""
	if evalContext.Rule.Message != "" {
		description += evalContext.Rule.Message
	}
	if evalContext.Error != nil {
		if description != "" {
			description += "\n"
		}
		description += "Error: " + evalContext.Error.Error()
	}
	if description != "" {
		alertJSON.SetPath([]string{"annotations", "description"}, description)
	}
	if evalContext.ImagePublicURL != "" {
		alertJSON.SetPath([]string{"annotations", "image"}, evalContext.ImagePublicURL)
	}

	// Labels (from metrics tags + AlertRuleTags + mandatory alertname).
	tags := make(map[string]string)
	if match != nil {
		if len(match.Tags) == 0 {
			tags["metric"] = match.Metric
		} else {
			for k, v := range match.Tags {
				tags[replaceIllegalCharsInLabelname(k)] = v
			}
		}
	}
	for _, tag := range evalContext.Rule.AlertRuleTags {
		tags[tag.Key] = tag.Value
	}
	tags["alertname"] = evalContext.Rule.Name
	alertJSON.Set("labels", tags)
	return alertJSON
}

// Notify sends alert notifications to the alert manager
func (am *AlertmanagerNotifier) Notify(evalContext *alerting.EvalContext) error {
	am.log.Info("Sending Alertmanager alert", "ruleId", evalContext.Rule.ID, "notification", am.Name)

	ruleURL, err := evalContext.GetRuleURL()
	if err != nil {
		am.log.Error("获取规则链接失败", "error", err)
		return err
	}

	// Send one alert per matching series.
	alerts := make([]interface{}, 0)
	for _, match := range evalContext.EvalMatches {
		alert := am.createAlert(evalContext, match, ruleURL)
		alerts = append(alerts, alert)
	}

	// This happens on ExecutionError or NoData
	if len(alerts) == 0 {
		alert := am.createAlert(evalContext, nil, ruleURL)
		alerts = append(alerts, alert)
	}

	bodyJSON := simplejson.NewFromAny(alerts)
	body, _ := bodyJSON.MarshalJSON()

	cmd := &models.SendWebhookSync{
		Url:        am.URL + "/api/v1/alerts",
		HttpMethod: "POST",
		Body:       string(body),
	}

	if err := bus.DispatchCtx(evalContext.Ctx, cmd); err != nil {
		am.log.Error("无法发送alertmanager", "error", err, "alertmanager", am.Name)
		return err
	}

	return nil
}

// regexp that matches all none valid label name characters
// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
var labelNamePattern = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func replaceIllegalCharsInLabelname(input string) string {
	return labelNamePattern.ReplaceAllString(input, "_")
}
