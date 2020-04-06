package models

import (
	"errors"
	"strings"
	"time"
)

// Typed errors
var (
	ErrFolderNotFound                = errors.New("找不到文件夹")
	ErrFolderVersionMismatch         = errors.New("该文件夹已被其他人更改")
	ErrFolderTitleEmpty              = errors.New("文件夹标题不能为空")
	ErrFolderWithSameUIDExists       = errors.New("已存在具有相同uid的文件夹/仪表板")
	ErrFolderSameNameExists          = errors.New("已存在具有相同名称的常规文件夹中的文件夹或仪表板")
	ErrFolderFailedGenerateUniqueUid = errors.New("无法生成唯一文件夹ID")
	ErrFolderAccessDenied            = errors.New("拒绝访问文件夹")
)

type Folder struct {
	Id      int64
	Uid     string
	Title   string
	Url     string
	Version int

	Created time.Time
	Updated time.Time

	UpdatedBy int64
	CreatedBy int64
	HasAcl    bool
}

// GetDashboardModel turns the command into the saveable model
func (cmd *CreateFolderCommand) GetDashboardModel(orgId int64, userId int64) *Dashboard {
	dashFolder := NewDashboardFolder(strings.TrimSpace(cmd.Title))
	dashFolder.OrgId = orgId
	dashFolder.SetUid(strings.TrimSpace(cmd.Uid))

	if userId == 0 {
		userId = -1
	}

	dashFolder.CreatedBy = userId
	dashFolder.UpdatedBy = userId
	dashFolder.UpdateSlug()

	return dashFolder
}

// UpdateDashboardModel updates an existing model from command into model for update
func (cmd *UpdateFolderCommand) UpdateDashboardModel(dashFolder *Dashboard, orgId int64, userId int64) {
	dashFolder.OrgId = orgId
	dashFolder.Title = strings.TrimSpace(cmd.Title)
	dashFolder.Data.Set("title", dashFolder.Title)

	if cmd.Uid != "" {
		dashFolder.SetUid(cmd.Uid)
	}

	dashFolder.SetVersion(cmd.Version)
	dashFolder.IsFolder = true

	if userId == 0 {
		userId = -1
	}

	dashFolder.UpdatedBy = userId
	dashFolder.UpdateSlug()
}

//
// COMMANDS
//

type CreateFolderCommand struct {
	Uid   string `json:"uid"`
	Title string `json:"title"`

	Result *Folder
}

type UpdateFolderCommand struct {
	Uid       string `json:"uid"`
	Title     string `json:"title"`
	Version   int    `json:"version"`
	Overwrite bool   `json:"overwrite"`

	Result *Folder
}

//
// QUERIES
//

type HasEditPermissionInFoldersQuery struct {
	SignedInUser *SignedInUser
	Result       bool
}
