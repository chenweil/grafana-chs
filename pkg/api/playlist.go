package api

import (
	"net/http"

	"github.com/grafana/grafana/pkg/bus"
	_ "github.com/grafana/grafana/pkg/infra/log"
	m "github.com/grafana/grafana/pkg/models"
)

func ValidateOrgPlaylist(c *m.ReqContext) {
	id := c.ParamsInt64(":id")
	query := m.GetPlaylistByIdQuery{Id: id}
	err := bus.Dispatch(&query)

	if err != nil {
		c.JsonApiErr(404, "找不到播放列表", err)
		return
	}

	if query.Result.OrgId == 0 {
		c.JsonApiErr(404, "找不到播放列表", err)
		return
	}

	if query.Result.OrgId != c.OrgId {
		c.JsonApiErr(403, "您无权编辑/查看播放列表", nil)
		return
	}

	items, itemsErr := LoadPlaylistItemDTOs(id)

	if itemsErr != nil {
		c.JsonApiErr(404, "找不到播放列表项目", err)
		return
	}

	if len(items) == 0 && c.Context.Req.Method != http.MethodDelete {
		c.JsonApiErr(404, "播放列表为空", itemsErr)
		return
	}
}

func SearchPlaylists(c *m.ReqContext) Response {
	query := c.Query("query")
	limit := c.QueryInt("limit")

	if limit == 0 {
		limit = 1000
	}

	searchQuery := m.GetPlaylistsQuery{
		Name:  query,
		Limit: limit,
		OrgId: c.OrgId,
	}

	err := bus.Dispatch(&searchQuery)
	if err != nil {
		return Error(500, "搜索失败", err)
	}

	return JSON(200, searchQuery.Result)
}

func GetPlaylist(c *m.ReqContext) Response {
	id := c.ParamsInt64(":id")
	cmd := m.GetPlaylistByIdQuery{Id: id}

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "找不到播放列表", err)
	}

	playlistDTOs, _ := LoadPlaylistItemDTOs(id)

	dto := &m.PlaylistDTO{
		Id:       cmd.Result.Id,
		Name:     cmd.Result.Name,
		Interval: cmd.Result.Interval,
		OrgId:    cmd.Result.OrgId,
		Items:    playlistDTOs,
	}

	return JSON(200, dto)
}

func LoadPlaylistItemDTOs(id int64) ([]m.PlaylistItemDTO, error) {
	playlistitems, err := LoadPlaylistItems(id)

	if err != nil {
		return nil, err
	}

	playlistDTOs := make([]m.PlaylistItemDTO, 0)

	for _, item := range playlistitems {
		playlistDTOs = append(playlistDTOs, m.PlaylistItemDTO{
			Id:         item.Id,
			PlaylistId: item.PlaylistId,
			Type:       item.Type,
			Value:      item.Value,
			Order:      item.Order,
			Title:      item.Title,
		})
	}

	return playlistDTOs, nil
}

func LoadPlaylistItems(id int64) ([]m.PlaylistItem, error) {
	itemQuery := m.GetPlaylistItemsByIdQuery{PlaylistId: id}
	if err := bus.Dispatch(&itemQuery); err != nil {
		return nil, err
	}

	return *itemQuery.Result, nil
}

func GetPlaylistItems(c *m.ReqContext) Response {
	id := c.ParamsInt64(":id")

	playlistDTOs, err := LoadPlaylistItemDTOs(id)

	if err != nil {
		return Error(500, "无法加载播放列表项", err)
	}

	return JSON(200, playlistDTOs)
}

func GetPlaylistDashboards(c *m.ReqContext) Response {
	playlistID := c.ParamsInt64(":id")

	playlists, err := LoadPlaylistDashboards(c.OrgId, c.SignedInUser, playlistID)
	if err != nil {
		return Error(500, "无法加载仪表板", err)
	}

	return JSON(200, playlists)
}

func DeletePlaylist(c *m.ReqContext) Response {
	id := c.ParamsInt64(":id")

	cmd := m.DeletePlaylistCommand{Id: id, OrgId: c.OrgId}
	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法删除播放列表", err)
	}

	return JSON(200, "")
}

func CreatePlaylist(c *m.ReqContext, cmd m.CreatePlaylistCommand) Response {
	cmd.OrgId = c.OrgId

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法创建播放列表", err)
	}

	return JSON(200, cmd.Result)
}

func UpdatePlaylist(c *m.ReqContext, cmd m.UpdatePlaylistCommand) Response {
	cmd.OrgId = c.OrgId
	cmd.Id = c.ParamsInt64(":id")

	if err := bus.Dispatch(&cmd); err != nil {
		return Error(500, "无法保存播放列表", err)
	}

	playlistDTOs, err := LoadPlaylistItemDTOs(cmd.Id)
	if err != nil {
		return Error(500, "无法保存播放列表", err)
	}

	cmd.Result.Items = playlistDTOs
	return JSON(200, cmd.Result)
}
