package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/metrics"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/guardian"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

var client = &http.Client{
	Timeout:   time.Second * 5,
	Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
}

func GetSharingOptions(c *m.ReqContext) {
	c.JSON(200, util.DynMap{
		"externalSnapshotURL":  setting.ExternalSnapshotUrl,
		"externalSnapshotName": setting.ExternalSnapshotName,
		"externalEnabled":      setting.ExternalEnabled,
	})
}

type CreateExternalSnapshotResponse struct {
	Key       string `json:"key"`
	DeleteKey string `json:"deleteKey"`
	Url       string `json:"url"`
	DeleteUrl string `json:"deleteUrl"`
}

func createExternalDashboardSnapshot(cmd m.CreateDashboardSnapshotCommand) (*CreateExternalSnapshotResponse, error) {
	var createSnapshotResponse CreateExternalSnapshotResponse
	message := map[string]interface{}{
		"name":      cmd.Name,
		"expires":   cmd.Expires,
		"dashboard": cmd.Dashboard,
	}

	messageBytes, err := simplejson.NewFromAny(message).Encode()
	if err != nil {
		return nil, err
	}

	response, err := client.Post(setting.ExternalSnapshotUrl+"/api/snapshots", "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("创建外部快照响应状态代码为 %d", response.StatusCode)
	}

	if err := json.NewDecoder(response.Body).Decode(&createSnapshotResponse); err != nil {
		return nil, err
	}

	return &createSnapshotResponse, nil
}

// POST /api/snapshots
func CreateDashboardSnapshot(c *m.ReqContext, cmd m.CreateDashboardSnapshotCommand) {
	if cmd.Name == "" {
		cmd.Name = "Unnamed snapshot"
	}

	var url string
	cmd.ExternalUrl = ""
	cmd.OrgId = c.OrgId
	cmd.UserId = c.UserId

	if cmd.External {
		if !setting.ExternalEnabled {
			c.JsonApiErr(403, "外部仪表板创建已禁用", nil)
			return
		}

		response, err := createExternalDashboardSnapshot(cmd)
		if err != nil {
			c.JsonApiErr(500, "无法创建外部快照", err)
			return
		}

		url = response.Url
		cmd.Key = response.Key
		cmd.DeleteKey = response.DeleteKey
		cmd.ExternalUrl = response.Url
		cmd.ExternalDeleteUrl = response.DeleteUrl
		cmd.Dashboard = simplejson.New()

		metrics.MApiDashboardSnapshotExternal.Inc()
	} else {
		if cmd.Key == "" {
			cmd.Key = util.GetRandomString(32)
		}

		if cmd.DeleteKey == "" {
			cmd.DeleteKey = util.GetRandomString(32)
		}

		url = setting.ToAbsUrl("dashboard/snapshot/" + cmd.Key)

		metrics.MApiDashboardSnapshotCreate.Inc()
	}

	if err := bus.Dispatch(&cmd); err != nil {
		c.JsonApiErr(500, "无法创建快照", err)
		return
	}

	c.JSON(200, util.DynMap{
		"key":       cmd.Key,
		"deleteKey": cmd.DeleteKey,
		"url":       url,
		"deleteUrl": setting.ToAbsUrl("api/snapshots-delete/" + cmd.DeleteKey),
	})
}

// GET /api/snapshots/:key
func GetDashboardSnapshot(c *m.ReqContext) {
	key := c.Params(":key")
	query := &m.GetDashboardSnapshotQuery{Key: key}

	err := bus.Dispatch(query)
	if err != nil {
		c.JsonApiErr(500, "无法获得仪表板快照", err)
		return
	}

	snapshot := query.Result

	// expired snapshots should also be removed from db
	if snapshot.Expires.Before(time.Now()) {
		c.JsonApiErr(404, "未找到仪表板快照", err)
		return
	}

	dto := dtos.DashboardFullWithMeta{
		Dashboard: snapshot.Dashboard,
		Meta: dtos.DashboardMeta{
			Type:       m.DashTypeSnapshot,
			IsSnapshot: true,
			Created:    snapshot.Created,
			Expires:    snapshot.Expires,
		},
	}

	metrics.MApiDashboardSnapshotGet.Inc()

	c.Resp.Header().Set("Cache-Control", "public, max-age=3600")
	c.JSON(200, dto)
}

func deleteExternalDashboardSnapshot(externalUrl string) error {
	response, err := client.Get(externalUrl)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		return nil
	}

	// Gracefully ignore "snapshot not found" errors as they could have already
	// been removed either via the cleanup script or by request.
	if response.StatusCode == 500 {
		var respJson map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&respJson); err != nil {
			return err
		}

		if respJson["message"] == "无法获得仪表板快照" {
			return nil
		}
	}

	return fmt.Errorf("删除外部快照时出现意外响应。 状态代码： %d", response.StatusCode)
}

// GET /api/snapshots-delete/:deleteKey
func DeleteDashboardSnapshotByDeleteKey(c *m.ReqContext) Response {
	key := c.Params(":deleteKey")

	query := &m.GetDashboardSnapshotQuery{DeleteKey: key}

	err := bus.Dispatch(query)
	if err != nil {
		return Error(500, "无法获得仪表板快照", err)
	}

	if query.Result.External {
		err := deleteExternalDashboardSnapshot(query.Result.ExternalDeleteUrl)
		if err != nil {
			return Error(500, "无法删除外部仪表板", err)
		}
	}

	cmd := &m.DeleteDashboardSnapshotCommand{DeleteKey: query.Result.DeleteKey}

	if err := bus.Dispatch(cmd); err != nil {
		return Error(500, "无法删除仪表板快照", err)
	}

	return JSON(200, util.DynMap{"message": "快照已删除。 它可能需要一个小时才能从任何CDN缓存中清除。"})
}

// DELETE /api/snapshots/:key
func DeleteDashboardSnapshot(c *m.ReqContext) Response {
	key := c.Params(":key")

	query := &m.GetDashboardSnapshotQuery{Key: key}

	err := bus.Dispatch(query)
	if err != nil {
		return Error(500, "无法获得仪表板快照", err)
	}

	if query.Result == nil {
		return Error(404, "无法获得仪表板快照", nil)
	}
	dashboard := query.Result.Dashboard
	dashboardID := dashboard.Get("id").MustInt64()

	guardian := guardian.New(dashboardID, c.OrgId, c.SignedInUser)
	canEdit, err := guardian.CanEdit()
	if err != nil {
		return Error(500, "检查快照权限时出错", err)
	}

	if !canEdit && query.Result.UserId != c.SignedInUser.UserId {
		return Error(403, "访问被拒绝此快照", nil)
	}

	if query.Result.External {
		err := deleteExternalDashboardSnapshot(query.Result.ExternalDeleteUrl)
		if err != nil {
			return Error(500, "无法删除外部仪表板", err)
		}
	}

	cmd := &m.DeleteDashboardSnapshotCommand{DeleteKey: query.Result.DeleteKey}

	if err := bus.Dispatch(cmd); err != nil {
		return Error(500, "无法删除仪表板快照", err)
	}

	return JSON(200, util.DynMap{"message": "快照已删除。 它可能需要一个小时才能从任何CDN缓存中清除。"})
}

// GET /api/dashboard/snapshots
func SearchDashboardSnapshots(c *m.ReqContext) Response {
	query := c.Query("query")
	limit := c.QueryInt("limit")

	if limit == 0 {
		limit = 1000
	}

	searchQuery := m.GetDashboardSnapshotsQuery{
		Name:         query,
		Limit:        limit,
		OrgId:        c.OrgId,
		SignedInUser: c.SignedInUser,
	}

	err := bus.Dispatch(&searchQuery)
	if err != nil {
		return Error(500, "搜索失败", err)
	}

	dtos := make([]*m.DashboardSnapshotDTO, len(searchQuery.Result))
	for i, snapshot := range searchQuery.Result {
		dtos[i] = &m.DashboardSnapshotDTO{
			Id:          snapshot.Id,
			Name:        snapshot.Name,
			Key:         snapshot.Key,
			OrgId:       snapshot.OrgId,
			UserId:      snapshot.UserId,
			External:    snapshot.External,
			ExternalUrl: snapshot.ExternalUrl,
			Expires:     snapshot.Expires,
			Created:     snapshot.Created,
			Updated:     snapshot.Updated,
		}
	}

	return JSON(200, dtos)
}
