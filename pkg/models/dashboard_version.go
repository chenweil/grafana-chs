package models

import (
	"errors"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
)

var (
	ErrDashboardVersionNotFound = errors.New("未找到仪表板版本")
	ErrNoVersionsForDashboardId = errors.New("找不到给定DashboardId的仪表板版本")
)

// A DashboardVersion represents the comparable data in a dashboard, allowing
// diffs of the dashboard to be performed.
type DashboardVersion struct {
	Id            int64 `json:"id"`
	DashboardId   int64 `json:"dashboardId"`
	ParentVersion int   `json:"parentVersion"`
	RestoredFrom  int   `json:"restoredFrom"`
	Version       int   `json:"version"`

	Created   time.Time `json:"created"`
	CreatedBy int64     `json:"createdBy"`

	Message string           `json:"message"`
	Data    *simplejson.Json `json:"data"`
}

// DashboardVersionMeta extends the dashboard version model with the names
// associated with the UserIds, overriding the field with the same name from
// the DashboardVersion model.
type DashboardVersionMeta struct {
	Id            int64            `json:"id"`
	DashboardId   int64            `json:"dashboardId"`
	ParentVersion int              `json:"parentVersion"`
	RestoredFrom  int              `json:"restoredFrom"`
	Version       int              `json:"version"`
	Created       time.Time        `json:"created"`
	Message       string           `json:"message"`
	Data          *simplejson.Json `json:"data"`
	CreatedBy     string           `json:"createdBy"`
}

// DashboardVersionDTO represents a dashboard version, without the dashboard
// map.
type DashboardVersionDTO struct {
	Id            int64     `json:"id"`
	DashboardId   int64     `json:"dashboardId"`
	ParentVersion int       `json:"parentVersion"`
	RestoredFrom  int       `json:"restoredFrom"`
	Version       int       `json:"version"`
	Created       time.Time `json:"created"`
	CreatedBy     string    `json:"createdBy"`
	Message       string    `json:"message"`
}

//
// Queries
//

type GetDashboardVersionQuery struct {
	DashboardId int64
	OrgId       int64
	Version     int

	Result *DashboardVersion
}

type GetDashboardVersionsQuery struct {
	DashboardId int64
	OrgId       int64
	Limit       int
	Start       int

	Result []*DashboardVersionDTO
}

//
// Commands
//

type DeleteExpiredVersionsCommand struct {
	DeletedRows int64
}
