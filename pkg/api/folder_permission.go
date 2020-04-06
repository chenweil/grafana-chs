package api

import (
	"time"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/dashboards"
	"github.com/grafana/grafana/pkg/services/guardian"
)

func GetFolderPermissionList(c *m.ReqContext) Response {
	s := dashboards.NewFolderService(c.OrgId, c.SignedInUser)
	folder, err := s.GetFolderByUID(c.Params(":uid"))

	if err != nil {
		return toFolderError(err)
	}

	g := guardian.New(folder.Id, c.OrgId, c.SignedInUser)

	if canAdmin, err := g.CanAdmin(); err != nil || !canAdmin {
		return toFolderError(m.ErrFolderAccessDenied)
	}

	acl, err := g.GetAcl()
	if err != nil {
		return Error(500, "无法获得文件夹权限", err)
	}

	for _, perm := range acl {
		perm.FolderId = folder.Id
		perm.DashboardId = 0

		perm.UserAvatarUrl = dtos.GetGravatarUrl(perm.UserEmail)

		if perm.TeamId > 0 {
			perm.TeamAvatarUrl = dtos.GetGravatarUrlWithDefault(perm.TeamEmail, perm.Team)
		}

		if perm.Slug != "" {
			perm.Url = m.GetDashboardFolderUrl(perm.IsFolder, perm.Uid, perm.Slug)
		}
	}

	return JSON(200, acl)
}

func UpdateFolderPermissions(c *m.ReqContext, apiCmd dtos.UpdateDashboardAclCommand) Response {
	s := dashboards.NewFolderService(c.OrgId, c.SignedInUser)
	folder, err := s.GetFolderByUID(c.Params(":uid"))

	if err != nil {
		return toFolderError(err)
	}

	g := guardian.New(folder.Id, c.OrgId, c.SignedInUser)
	canAdmin, err := g.CanAdmin()
	if err != nil {
		return toFolderError(err)
	}

	if !canAdmin {
		return toFolderError(m.ErrFolderAccessDenied)
	}

	cmd := m.UpdateDashboardAclCommand{}
	cmd.DashboardId = folder.Id

	for _, item := range apiCmd.Items {
		cmd.Items = append(cmd.Items, &m.DashboardAcl{
			OrgId:       c.OrgId,
			DashboardId: folder.Id,
			UserId:      item.UserId,
			TeamId:      item.TeamId,
			Role:        item.Role,
			Permission:  item.Permission,
			Created:     time.Now(),
			Updated:     time.Now(),
		})
	}

	if okToUpdate, err := g.CheckPermissionBeforeUpdate(m.PERMISSION_ADMIN, cmd.Items); err != nil || !okToUpdate {
		if err != nil {
			if err == guardian.ErrGuardianPermissionExists ||
				err == guardian.ErrGuardianOverride {
				return Error(400, err.Error(), err)
			}

			return Error(500, "检查文件夹权限时出错", err)
		}

		return Error(403, "无法删除文件夹的管理员权限", nil)
	}

	if err := bus.Dispatch(&cmd); err != nil {
		if err == m.ErrDashboardAclInfoMissing {
			err = m.ErrFolderAclInfoMissing
		}
		if err == m.ErrDashboardPermissionDashboardEmpty {
			err = m.ErrFolderPermissionFolderEmpty
		}

		if err == m.ErrFolderAclInfoMissing || err == m.ErrFolderPermissionFolderEmpty {
			return Error(409, err.Error(), err)
		}

		return Error(500, "无法创建权限", err)
	}

	return Success("文件夹权限已更新")
}
