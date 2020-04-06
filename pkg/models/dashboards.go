package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/setting"
)

// Typed errors
var (
	ErrDashboardNotFound                         = errors.New("未找到仪表板")
	ErrDashboardFolderNotFound                   = errors.New("找不到文件夹")
	ErrDashboardSnapshotNotFound                 = errors.New("未找到仪表板快照")
	ErrDashboardWithSameUIDExists                = errors.New("已存在具有相同uid的仪表板")
	ErrDashboardWithSameNameInFolderExists       = errors.New("已存在文件夹中具有相同名称的仪表板")
	ErrDashboardVersionMismatch                  = errors.New("仪表板已被其他人更改")
	ErrDashboardTitleEmpty                       = errors.New("仪表板标题不能为空")
	ErrDashboardFolderCannotHaveParent           = errors.New("无法将仪表板文件夹添加到其他文件夹")
	ErrDashboardsWithSameSlugExists              = errors.New("存在具有相同段塞的多个仪表板")
	ErrDashboardFailedGenerateUniqueUid          = errors.New("无法生成唯一的仪表板ID")
	ErrDashboardTypeMismatch                     = errors.New("仪表板无法更改为文件夹")
	ErrDashboardFolderWithSameNameAsDashboard    = errors.New("文件夹名称不能与其中一个仪表板相同")
	ErrDashboardWithSameNameAsFolder             = errors.New("仪表板名称不能与文件夹相同")
	ErrDashboardFolderNameExists                 = errors.New("已存在具有该名称的文件夹")
	ErrDashboardUpdateAccessDenied               = errors.New("访问被拒绝以保存仪表板")
	ErrDashboardInvalidUid                       = errors.New("uid包含非法字符")
	ErrDashboardUidToLong                        = errors.New("uid太长，最多40个字符")
	ErrDashboardCannotSaveProvisionedDashboard   = errors.New("无法保存配置的仪表板")
	ErrDashboardCannotDeleteProvisionedDashboard = errors.New("配置的仪表板无法删除")
	RootFolderName                               = "默认"
)

type UpdatePluginDashboardError struct {
	PluginId string
}

func (d UpdatePluginDashboardError) Error() string {
	return "Dashboard belong to plugin"
}

var (
	DashTypeJson     = "file"
	DashTypeDB       = "db"
	DashTypeScript   = "script"
	DashTypeSnapshot = "snapshot"
)

// Dashboard model
type Dashboard struct {
	Id       int64
	Uid      string
	Slug     string
	OrgId    int64
	GnetId   int64
	Version  int
	PluginId string

	Created time.Time
	Updated time.Time

	UpdatedBy int64
	CreatedBy int64
	FolderId  int64
	IsFolder  bool
	HasAcl    bool

	Title string
	Data  *simplejson.Json
}

func (d *Dashboard) SetId(id int64) {
	d.Id = id
	d.Data.Set("id", id)
}

func (d *Dashboard) SetUid(uid string) {
	d.Uid = uid
	d.Data.Set("uid", uid)
}

func (d *Dashboard) SetVersion(version int) {
	d.Version = version
	d.Data.Set("version", version)
}

// GetDashboardIdForSavePermissionCheck return the dashboard id to be used for checking permission of dashboard
func (d *Dashboard) GetDashboardIdForSavePermissionCheck() int64 {
	if d.Id == 0 {
		return d.FolderId
	}

	return d.Id
}

// NewDashboard creates a new dashboard
func NewDashboard(title string) *Dashboard {
	dash := &Dashboard{}
	dash.Data = simplejson.New()
	dash.Data.Set("title", title)
	dash.Title = title
	dash.Created = time.Now()
	dash.Updated = time.Now()
	dash.UpdateSlug()
	return dash
}

// NewDashboardFolder creates a new dashboard folder
func NewDashboardFolder(title string) *Dashboard {
	folder := NewDashboard(title)
	folder.IsFolder = true
	folder.Data.Set("schemaVersion", 17)
	folder.Data.Set("version", 0)
	folder.IsFolder = true
	return folder
}

// GetTags turns the tags in data json into go string array
func (dash *Dashboard) GetTags() []string {
	return dash.Data.Get("tags").MustStringArray()
}

func NewDashboardFromJson(data *simplejson.Json) *Dashboard {
	dash := &Dashboard{}
	dash.Data = data
	dash.Title = dash.Data.Get("title").MustString()
	dash.UpdateSlug()
	update := false

	if id, err := dash.Data.Get("id").Float64(); err == nil {
		dash.Id = int64(id)
		update = true
	}

	if uid, err := dash.Data.Get("uid").String(); err == nil {
		dash.Uid = uid
		update = true
	}

	if version, err := dash.Data.Get("version").Float64(); err == nil && update {
		dash.Version = int(version)
		dash.Updated = time.Now()
	} else {
		dash.Data.Set("version", 0)
		dash.Created = time.Now()
		dash.Updated = time.Now()
	}

	if gnetId, err := dash.Data.Get("gnetId").Float64(); err == nil {
		dash.GnetId = int64(gnetId)
	}

	return dash
}

// GetDashboardModel turns the command into the saveable model
func (cmd *SaveDashboardCommand) GetDashboardModel() *Dashboard {
	dash := NewDashboardFromJson(cmd.Dashboard)
	userId := cmd.UserId

	if userId == 0 {
		userId = -1
	}

	dash.UpdatedBy = userId
	dash.OrgId = cmd.OrgId
	dash.PluginId = cmd.PluginId
	dash.IsFolder = cmd.IsFolder
	dash.FolderId = cmd.FolderId
	dash.UpdateSlug()
	return dash
}

// GetString a
func (dash *Dashboard) GetString(prop string, defaultValue string) string {
	return dash.Data.Get(prop).MustString(defaultValue)
}

// UpdateSlug updates the slug
func (dash *Dashboard) UpdateSlug() {
	title := dash.Data.Get("title").MustString()
	dash.Slug = SlugifyTitle(title)
}

func SlugifyTitle(title string) string {
	return slug.Make(strings.ToLower(title))
}

// GetUrl return the html url for a folder if it's folder, otherwise for a dashboard
func (dash *Dashboard) GetUrl() string {
	return GetDashboardFolderUrl(dash.IsFolder, dash.Uid, dash.Slug)
}

// Return the html url for a dashboard
func (dash *Dashboard) GenerateUrl() string {
	return GetDashboardUrl(dash.Uid, dash.Slug)
}

// GetDashboardFolderUrl return the html url for a folder if it's folder, otherwise for a dashboard
func GetDashboardFolderUrl(isFolder bool, uid string, slug string) string {
	if isFolder {
		return GetFolderUrl(uid, slug)
	}

	return GetDashboardUrl(uid, slug)
}

// GetDashboardUrl return the html url for a dashboard
func GetDashboardUrl(uid string, slug string) string {
	return fmt.Sprintf("%s/d/%s/%s", setting.AppSubUrl, uid, slug)
}

// GetFullDashboardUrl return the full url for a dashboard
func GetFullDashboardUrl(uid string, slug string) string {
	return fmt.Sprintf("%sd/%s/%s", setting.AppUrl, uid, slug)
}

// GetFolderUrl return the html url for a folder
func GetFolderUrl(folderUid string, slug string) string {
	return fmt.Sprintf("%s/dashboards/f/%s/%s", setting.AppSubUrl, folderUid, slug)
}

type ValidateDashboardBeforeSaveResult struct {
	IsParentFolderChanged bool
}

//
// COMMANDS
//

type SaveDashboardCommand struct {
	Dashboard    *simplejson.Json `json:"dashboard" binding:"Required"`
	UserId       int64            `json:"userId"`
	Overwrite    bool             `json:"overwrite"`
	Message      string           `json:"message"`
	OrgId        int64            `json:"-"`
	RestoredFrom int              `json:"-"`
	PluginId     string           `json:"-"`
	FolderId     int64            `json:"folderId"`
	IsFolder     bool             `json:"isFolder"`

	UpdatedAt time.Time

	Result *Dashboard
}

type DashboardProvisioning struct {
	Id          int64
	DashboardId int64
	Name        string
	ExternalId  string
	CheckSum    string
	Updated     int64
}

type SaveProvisionedDashboardCommand struct {
	DashboardCmd          *SaveDashboardCommand
	DashboardProvisioning *DashboardProvisioning

	Result *Dashboard
}

type DeleteDashboardCommand struct {
	Id    int64
	OrgId int64
}

type ValidateDashboardBeforeSaveCommand struct {
	OrgId     int64
	Dashboard *Dashboard
	Overwrite bool
	Result    *ValidateDashboardBeforeSaveResult
}

//
// QUERIES
//

type GetDashboardQuery struct {
	Slug  string // required if no Id or Uid is specified
	Id    int64  // optional if slug is set
	Uid   string // optional if slug is set
	OrgId int64

	Result *Dashboard
}

type DashboardTagCloudItem struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

type GetDashboardTagsQuery struct {
	OrgId  int64
	Result []*DashboardTagCloudItem
}

type GetDashboardsQuery struct {
	DashboardIds []int64
	Result       []*Dashboard
}

type GetDashboardPermissionsForUserQuery struct {
	DashboardIds []int64
	OrgId        int64
	UserId       int64
	OrgRole      RoleType
	Result       []*DashboardPermissionForUser
}

type GetDashboardsByPluginIdQuery struct {
	OrgId    int64
	PluginId string
	Result   []*Dashboard
}

type GetDashboardSlugByIdQuery struct {
	Id     int64
	Result string
}

type GetProvisionedDashboardDataByIdQuery struct {
	DashboardId int64
	Result      *DashboardProvisioning
}

type GetProvisionedDashboardDataQuery struct {
	Name   string
	Result []*DashboardProvisioning
}

type GetDashboardsBySlugQuery struct {
	OrgId int64
	Slug  string

	Result []*Dashboard
}

type DashboardPermissionForUser struct {
	DashboardId    int64          `json:"dashboardId"`
	Permission     PermissionType `json:"permission"`
	PermissionName string         `json:"permissionName"`
}

type DashboardRef struct {
	Uid  string
	Slug string
}

type GetDashboardRefByIdQuery struct {
	Id     int64
	Result *DashboardRef
}

type UnprovisionDashboardCommand struct {
	Id int64
}
