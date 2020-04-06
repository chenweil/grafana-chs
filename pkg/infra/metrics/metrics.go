package metrics

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/grafana/grafana/pkg/setting"
)

const exporterName = "grafana"

var (
	// MInstanceStart is a metric counter for started instances
	MInstanceStart prometheus.Counter

	// MPageStatus is a metric page http response status
	MPageStatus *prometheus.CounterVec

	// MApiStatus is a metric api http response status
	MApiStatus *prometheus.CounterVec

	// MProxyStatus is a metric proxy http response status
	MProxyStatus *prometheus.CounterVec

	// MHttpRequestTotal is a metric http request counter
	MHttpRequestTotal *prometheus.CounterVec

	// MHttpRequestSummary is a metric http request summary
	MHttpRequestSummary *prometheus.SummaryVec

	// MApiUserSignUpStarted is a metric amount of users who started the signup flow
	MApiUserSignUpStarted prometheus.Counter

	// MApiUserSignUpCompleted is a metric amount of users who completed the signup flow
	MApiUserSignUpCompleted prometheus.Counter

	// MApiUserSignUpInvite is a metric amount of users who have been invited
	MApiUserSignUpInvite prometheus.Counter

	// MApiDashboardSave is a metric summary for dashboard save duration
	MApiDashboardSave prometheus.Summary

	// MApiDashboardGet is a metric summary for dashboard get duration
	MApiDashboardGet prometheus.Summary

	// MApiDashboardSearch is a metric summary for dashboard search duration
	MApiDashboardSearch prometheus.Summary

	// MApiAdminUserCreate is a metric api admin user created counter
	MApiAdminUserCreate prometheus.Counter

	// MApiLoginPost is a metric api login post counter
	MApiLoginPost prometheus.Counter

	// MApiLoginOAuth is a metric api login oauth counter
	MApiLoginOAuth prometheus.Counter

	// MApiLoginSAML is a metric api login SAML counter
	MApiLoginSAML prometheus.Counter

	// MApiOrgCreate is a metric api org created counter
	MApiOrgCreate prometheus.Counter

	// MApiDashboardSnapshotCreate is a metric dashboard snapshots created
	MApiDashboardSnapshotCreate prometheus.Counter

	// MApiDashboardSnapshotExternal is a metric external dashboard snapshots created
	MApiDashboardSnapshotExternal prometheus.Counter

	// MApiDashboardSnapshotGet is a metric loaded dashboards
	MApiDashboardSnapshotGet prometheus.Counter

	// MApiDashboardInsert is a metric dashboards inserted
	MApiDashboardInsert prometheus.Counter

	// MAlertingResultState is a metric alert execution result counter
	MAlertingResultState *prometheus.CounterVec

	// MAlertingNotificationSent is a metric counter for how many alert notifications been sent
	MAlertingNotificationSent *prometheus.CounterVec

	// MAwsCloudWatchGetMetricStatistics is a metric counter for getting metric statistics from aws
	MAwsCloudWatchGetMetricStatistics prometheus.Counter

	// MAwsCloudWatchListMetrics is a metric counter for getting list of metrics from aws
	MAwsCloudWatchListMetrics prometheus.Counter

	// MAwsCloudWatchGetMetricData is a metric counter for getting metric data time series from aws
	MAwsCloudWatchGetMetricData prometheus.Counter

	// MDBDataSourceQueryByID is a metric counter for getting datasource by id
	MDBDataSourceQueryByID prometheus.Counter

	// LDAPUsersSyncExecutionTime is a metric summary for LDAP users sync execution duration
	LDAPUsersSyncExecutionTime prometheus.Summary
)

// Timers
var (
	// MDataSourceProxyReqTimer is a metric summary for dataproxy request duration
	MDataSourceProxyReqTimer prometheus.Summary

	// MAlertingExecutionTime is a metric summary of alert exeuction duration
	MAlertingExecutionTime prometheus.Summary
)

// StatTotals
var (
	// MAlertingActiveAlerts is a metric amount of active alerts
	MAlertingActiveAlerts prometheus.Gauge

	// MStatTotalDashboards is a metric total amount of dashboards
	MStatTotalDashboards prometheus.Gauge

	// MStatTotalUsers is a metric total amount of users
	MStatTotalUsers prometheus.Gauge

	// MStatActiveUsers is a metric number of active users
	MStatActiveUsers prometheus.Gauge

	// MStatTotalOrgs is a metric total amount of orgs
	MStatTotalOrgs prometheus.Gauge

	// MStatTotalPlaylists is a metric total amount of playlists
	MStatTotalPlaylists prometheus.Gauge

	// StatsTotalViewers is a metric total amount of viewers
	StatsTotalViewers prometheus.Gauge

	// StatsTotalEditors is a metric total amount of editors
	StatsTotalEditors prometheus.Gauge

	// StatsTotalAdmins is a metric total amount of admins
	StatsTotalAdmins prometheus.Gauge

	// StatsTotalActiveViewers is a metric total amount of viewers
	StatsTotalActiveViewers prometheus.Gauge

	// StatsTotalActiveEditors is a metric total amount of active editors
	StatsTotalActiveEditors prometheus.Gauge

	// StatsTotalActiveAdmins is a metric total amount of active admins
	StatsTotalActiveAdmins prometheus.Gauge

	// grafanaBuildVersion is a metric with a constant '1' value labeled by version, revision, branch, and goversion from which Grafana was built
	grafanaBuildVersion *prometheus.GaugeVec
)

func init() {
	httpStatusCodes := []string{"200", "404", "500", "unknown"}
	MInstanceStart = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "instance_start_total",
		Help:      "启动实例的计数器",
		Namespace: exporterName,
	})

	MPageStatus = newCounterVecStartingAtZero(
		prometheus.CounterOpts{
			Name:      "page_response_status_total",
			Help:      "页面http响应状态",
			Namespace: exporterName,
		}, []string{"code"}, httpStatusCodes...)

	MApiStatus = newCounterVecStartingAtZero(
		prometheus.CounterOpts{
			Name:      "api_response_status_total",
			Help:      "api http响应状态",
			Namespace: exporterName,
		}, []string{"code"}, httpStatusCodes...)

	MProxyStatus = newCounterVecStartingAtZero(
		prometheus.CounterOpts{
			Name:      "proxy_response_status_total",
			Help:      "代理http响应状态",
			Namespace: exporterName,
		}, []string{"code"}, httpStatusCodes...)

	MHttpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "http请求计数器",
		},
		[]string{"handler", "statuscode", "method"},
	)

	MHttpRequestSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_duration_milliseconds",
			Help: "http请求摘要",
		},
		[]string{"handler", "statuscode", "method"},
	)

	MApiUserSignUpStarted = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_user_signup_started_total",
		Help:      "启动注册流程的用户数量",
		Namespace: exporterName,
	})

	MApiUserSignUpCompleted = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_user_signup_completed_total",
		Help:      "完成注册流程的用户数量",
		Namespace: exporterName,
	})

	MApiUserSignUpInvite = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_user_signup_invite_total",
		Help:      "被邀请的用户数量",
		Namespace: exporterName,
	})

	MApiDashboardSave = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "api_dashboard_save_milliseconds",
		Help:      "仪表板保存持续时间的摘要",
		Namespace: exporterName,
	})

	MApiDashboardGet = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "api_dashboard_get_milliseconds",
		Help:      "仪表板摘要持续时间",
		Namespace: exporterName,
	})

	MApiDashboardSearch = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "api_dashboard_search_milliseconds",
		Help:      "仪表板搜索持续时间摘要",
		Namespace: exporterName,
	})

	MApiAdminUserCreate = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_admin_user_created_total",
		Help:      "api admin用户创建了计数器",
		Namespace: exporterName,
	})

	MApiLoginPost = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_login_post_total",
		Help:      "api登录计数器",
		Namespace: exporterName,
	})

	MApiLoginOAuth = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_login_oauth_total",
		Help:      "api登录oauth计数器",
		Namespace: exporterName,
	})

	MApiLoginSAML = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_login_saml_total",
		Help:      "api登录saml计数器",
		Namespace: exporterName,
	})

	MApiOrgCreate = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_org_create_total",
		Help:      "api 团队创建计数器",
		Namespace: exporterName,
	})

	MApiDashboardSnapshotCreate = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_dashboard_snapshot_create_total",
		Help:      "创建的仪表板快照",
		Namespace: exporterName,
	})

	MApiDashboardSnapshotExternal = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_dashboard_snapshot_external_total",
		Help:      "创建外部仪表板快照",
		Namespace: exporterName,
	})

	MApiDashboardSnapshotGet = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_dashboard_snapshot_get_total",
		Help:      "加载仪表板",
		Namespace: exporterName,
	})

	MApiDashboardInsert = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "api_models_dashboard_insert_total",
		Help:      "仪表板已插入",
		Namespace: exporterName,
	})

	MAlertingResultState = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "alerting_result_total",
		Help:      "警报执行结果计数器",
		Namespace: exporterName,
	}, []string{"state"})

	MAlertingNotificationSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:      "alerting_notification_sent_total",
		Help:      "计数器已发送了多少个警报通知",
		Namespace: exporterName,
	}, []string{"type"})

	MAwsCloudWatchGetMetricStatistics = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "aws_cloudwatch_get_metric_statistics_total",
		Help:      "从aws获取度量统计信息的计数器",
		Namespace: exporterName,
	})

	MAwsCloudWatchListMetrics = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "aws_cloudwatch_list_metrics_total",
		Help:      "从aws获取指标列表的计数器",
		Namespace: exporterName,
	})

	MAwsCloudWatchGetMetricData = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "aws_cloudwatch_get_metric_data_total",
		Help:      "用于从aws获取度量数据时间序列的计数器",
		Namespace: exporterName,
	})

	MDBDataSourceQueryByID = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "db_datasource_query_by_id_total",
		Help:      "通过id获取数据源的计数器",
		Namespace: exporterName,
	})

	LDAPUsersSyncExecutionTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "ldap_users_sync_execution_time",
		Help:      "LDAP用户的摘要同步执行持续时间",
		Namespace: exporterName,
	})

	MDataSourceProxyReqTimer = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "api_dataproxy_request_all_milliseconds",
		Help:      "dataproxy请求持续时间的摘要",
		Namespace: exporterName,
	})

	MAlertingExecutionTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "alerting_execution_time_milliseconds",
		Help:      "警报执行持续时间的摘要",
		Namespace: exporterName,
	})

	MAlertingActiveAlerts = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "alerting_active_alerts",
		Help:      "活跃警报的数量",
		Namespace: exporterName,
	})

	MStatTotalDashboards = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_dashboard",
		Help:      "仪表板总数",
		Namespace: exporterName,
	})

	MStatTotalUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_total_users",
		Help:      "用户总数",
		Namespace: exporterName,
	})

	MStatActiveUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_active_users",
		Help:      "活跃用户数",
		Namespace: exporterName,
	})

	MStatTotalOrgs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_total_orgs",
		Help:      "团队总数",
		Namespace: exporterName,
	})

	MStatTotalPlaylists = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_total_playlists",
		Help:      "播放列表的总数",
		Namespace: exporterName,
	})

	StatsTotalViewers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_viewers",
		Help:      "观众总数",
		Namespace: exporterName,
	})

	StatsTotalEditors = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_editors",
		Help:      "编辑总数",
		Namespace: exporterName,
	})

	StatsTotalAdmins = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_admins",
		Help:      "管理员总数",
		Namespace: exporterName,
	})

	StatsTotalActiveViewers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_active_viewers",
		Help:      "观众总数",
		Namespace: exporterName,
	})

	StatsTotalActiveEditors = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_active_editors",
		Help:      "活动编辑总数",
		Namespace: exporterName,
	})

	StatsTotalActiveAdmins = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "stat_totals_active_admins",
		Help:      "活动管理员总数",
		Namespace: exporterName,
	})

	grafanaBuildVersion = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "build_info",
		Help:      "具有常量“1”值的度量标准，由Grafana构建的版本，修订版，分支和goversion标记",
		Namespace: exporterName,
	}, []string{"version", "revision", "branch", "goversion", "edition"})
}

// SetBuildInformation sets the build information for this binary
func SetBuildInformation(version, revision, branch string) {
	edition := "oss"
	if setting.IsEnterprise {
		edition = "enterprise"
	}

	grafanaBuildVersion.WithLabelValues(version, revision, branch, runtime.Version(), edition).Set(1)
}

func initMetricVars() {
	prometheus.MustRegister(
		MInstanceStart,
		MPageStatus,
		MApiStatus,
		MProxyStatus,
		MHttpRequestTotal,
		MHttpRequestSummary,
		MApiUserSignUpStarted,
		MApiUserSignUpCompleted,
		MApiUserSignUpInvite,
		MApiDashboardSave,
		MApiDashboardGet,
		MApiDashboardSearch,
		MDataSourceProxyReqTimer,
		MAlertingExecutionTime,
		MApiAdminUserCreate,
		MApiLoginPost,
		MApiLoginOAuth,
		MApiLoginSAML,
		MApiOrgCreate,
		MApiDashboardSnapshotCreate,
		MApiDashboardSnapshotExternal,
		MApiDashboardSnapshotGet,
		MApiDashboardInsert,
		MAlertingResultState,
		MAlertingNotificationSent,
		MAwsCloudWatchGetMetricStatistics,
		MAwsCloudWatchListMetrics,
		MAwsCloudWatchGetMetricData,
		MDBDataSourceQueryByID,
		LDAPUsersSyncExecutionTime,
		MAlertingActiveAlerts,
		MStatTotalDashboards,
		MStatTotalUsers,
		MStatActiveUsers,
		MStatTotalOrgs,
		MStatTotalPlaylists,
		StatsTotalViewers,
		StatsTotalEditors,
		StatsTotalAdmins,
		StatsTotalActiveViewers,
		StatsTotalActiveEditors,
		StatsTotalActiveAdmins,
		grafanaBuildVersion,
	)

}

func newCounterVecStartingAtZero(opts prometheus.CounterOpts, labels []string, labelValues ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(opts, labels)

	for _, label := range labelValues {
		counter.WithLabelValues(label).Add(0)
	}

	return counter
}

func newCounterStartingAtZero(opts prometheus.CounterOpts, labelValues ...string) prometheus.Counter {
	counter := prometheus.NewCounter(opts)
	counter.Add(0)

	return counter
}
