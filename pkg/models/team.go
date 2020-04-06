package models

import (
	"errors"
	"time"
)

// Typed errors
var (
	ErrTeamNotFound                         = errors.New("团队未找到")
	ErrTeamNameTaken                        = errors.New("团队名称被使用")
	ErrTeamMemberNotFound                   = errors.New("未找到团队成员")
	ErrLastTeamAdmin                        = errors.New("不允许删除上一个管理员")
	ErrNotAllowedToUpdateTeam               = errors.New("用户不允许更新团队")
	ErrNotAllowedToUpdateTeamInDifferentOrg = errors.New("用户不允许在另一个团队中更新团队")
)

// Team model
type Team struct {
	Id    int64  `json:"id"`
	OrgId int64  `json:"orgId"`
	Name  string `json:"name"`
	Email string `json:"email"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// ---------------------
// COMMANDS

type CreateTeamCommand struct {
	Name  string `json:"name" binding:"Required"`
	Email string `json:"email"`
	OrgId int64  `json:"-"`

	Result Team `json:"-"`
}

type UpdateTeamCommand struct {
	Id    int64
	Name  string
	Email string
	OrgId int64 `json:"-"`
}

type DeleteTeamCommand struct {
	OrgId int64
	Id    int64
}

type GetTeamByIdQuery struct {
	OrgId  int64
	Id     int64
	Result *TeamDTO
}

type GetTeamsByUserQuery struct {
	OrgId  int64
	UserId int64      `json:"userId"`
	Result []*TeamDTO `json:"teams"`
}

type SearchTeamsQuery struct {
	Query        string
	Name         string
	Limit        int
	Page         int
	OrgId        int64
	UserIdFilter int64

	Result SearchTeamQueryResult
}

type TeamDTO struct {
	Id          int64          `json:"id"`
	OrgId       int64          `json:"orgId"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	AvatarUrl   string         `json:"avatarUrl"`
	MemberCount int64          `json:"memberCount"`
	Permission  PermissionType `json:"permission"`
}

type SearchTeamQueryResult struct {
	TotalCount int64      `json:"totalCount"`
	Teams      []*TeamDTO `json:"teams"`
	Page       int        `json:"page"`
	PerPage    int        `json:"perPage"`
}
