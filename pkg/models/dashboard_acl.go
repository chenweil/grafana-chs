package models

import (
	"errors"
	"time"
)

type PermissionType int

const (
	PERMISSION_VIEW PermissionType = 1 << iota
	PERMISSION_EDIT
	PERMISSION_ADMIN
)

func (p PermissionType) String() string {
	names := map[int]string{
		int(PERMISSION_VIEW):  "View",
		int(PERMISSION_EDIT):  "Edit",
		int(PERMISSION_ADMIN): "Admin",
	}
	return names[int(p)]
}

// Typed errors
var (
	ErrDashboardAclInfoMissing           = errors.New("对于仪表板权限，用户ID和团队ID不能都为空")
	ErrDashboardPermissionDashboardEmpty = errors.New("对于仪表板权限，仪表板ID必须大于零")
	ErrFolderAclInfoMissing              = errors.New("对于文件夹权限，用户ID和团队ID不能都为空")
	ErrFolderPermissionFolderEmpty       = errors.New("对于文件夹权限，文件夹ID必须大于零")
)

// Dashboard ACL model
type DashboardAcl struct {
	Id          int64
	OrgId       int64
	DashboardId int64

	UserId     int64
	TeamId     int64
	Role       *RoleType // pointer to be nullable
	Permission PermissionType

	Created time.Time
	Updated time.Time
}

type DashboardAclInfoDTO struct {
	OrgId       int64 `json:"-"`
	DashboardId int64 `json:"dashboardId,omitempty"`
	FolderId    int64 `json:"folderId,omitempty"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	UserId         int64          `json:"userId"`
	UserLogin      string         `json:"userLogin"`
	UserEmail      string         `json:"userEmail"`
	UserAvatarUrl  string         `json:"userAvatarUrl"`
	TeamId         int64          `json:"teamId"`
	TeamEmail      string         `json:"teamEmail"`
	TeamAvatarUrl  string         `json:"teamAvatarUrl"`
	Team           string         `json:"team"`
	Role           *RoleType      `json:"role,omitempty"`
	Permission     PermissionType `json:"permission"`
	PermissionName string         `json:"permissionName"`
	Uid            string         `json:"uid"`
	Title          string         `json:"title"`
	Slug           string         `json:"slug"`
	IsFolder       bool           `json:"isFolder"`
	Url            string         `json:"url"`
	Inherited      bool           `json:"inherited"`
}

func (dto *DashboardAclInfoDTO) hasSameRoleAs(other *DashboardAclInfoDTO) bool {
	if dto.Role == nil || other.Role == nil {
		return false
	}

	return dto.UserId <= 0 && dto.TeamId <= 0 && dto.UserId == other.UserId && dto.TeamId == other.TeamId && *dto.Role == *other.Role
}

func (dto *DashboardAclInfoDTO) hasSameUserAs(other *DashboardAclInfoDTO) bool {
	return dto.UserId > 0 && dto.UserId == other.UserId
}

func (dto *DashboardAclInfoDTO) hasSameTeamAs(other *DashboardAclInfoDTO) bool {
	return dto.TeamId > 0 && dto.TeamId == other.TeamId
}

// IsDuplicateOf returns true if other item has same role, same user or same team
func (dto *DashboardAclInfoDTO) IsDuplicateOf(other *DashboardAclInfoDTO) bool {
	return dto.hasSameRoleAs(other) || dto.hasSameUserAs(other) || dto.hasSameTeamAs(other)
}

//
// COMMANDS
//

type UpdateDashboardAclCommand struct {
	DashboardId int64
	Items       []*DashboardAcl
}

//
// QUERIES
//
type GetDashboardAclInfoListQuery struct {
	DashboardId int64
	OrgId       int64
	Result      []*DashboardAclInfoDTO
}
