package cleanup

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/serverlock"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/setting"
)

type CleanUpService struct {
	log               log.Logger
	Cfg               *setting.Cfg                  `inject:""`
	ServerLockService *serverlock.ServerLockService `inject:""`
}

func init() {
	registry.RegisterService(&CleanUpService{})
}

func (srv *CleanUpService) Init() error {
	srv.log = log.New("cleanup")
	return nil
}

func (srv *CleanUpService) Run(ctx context.Context) error {
	srv.cleanUpTmpFiles()

	ticker := time.NewTicker(time.Minute * 10)
	for {
		select {
		case <-ticker.C:
			srv.cleanUpTmpFiles()
			srv.deleteExpiredSnapshots()
			srv.deleteExpiredDashboardVersions()
			srv.ServerLockService.LockAndExecute(ctx, "delete old login attempts", time.Minute*10, func() {
				srv.deleteOldLoginAttempts()
			})

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *CleanUpService) cleanUpTmpFiles() {
	if _, err := os.Stat(srv.Cfg.ImagesDir); os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir(srv.Cfg.ImagesDir)
	if err != nil {
		srv.log.Error("读取图片目录", "error", err)
		return
	}

	var toDelete []os.FileInfo
	var now = time.Now()

	for _, file := range files {
		if srv.shouldCleanupTempFile(file.ModTime(), now) {
			toDelete = append(toDelete, file)
		}
	}

	for _, file := range toDelete {
		fullPath := path.Join(srv.Cfg.ImagesDir, file.Name())
		err := os.Remove(fullPath)
		if err != nil {
			srv.log.Error("无法删除临时文件", "file", file.Name(), "error", err)
		}
	}

	srv.log.Debug("找到要删除的旧渲染图像", "deleted", len(toDelete), "kept", len(files))
}

func (srv *CleanUpService) shouldCleanupTempFile(filemtime time.Time, now time.Time) bool {
	if srv.Cfg.TempDataLifetime == 0 {
		return false
	}

	return filemtime.Add(srv.Cfg.TempDataLifetime).Before(now)
}

func (srv *CleanUpService) deleteExpiredSnapshots() {
	cmd := m.DeleteExpiredSnapshotsCommand{}
	if err := bus.Dispatch(&cmd); err != nil {
		srv.log.Error("无法删除过期的快照", "error", err.Error())
	} else {
		srv.log.Debug("删除过期的快照", "rows affected", cmd.DeletedRows)
	}
}

func (srv *CleanUpService) deleteExpiredDashboardVersions() {
	cmd := m.DeleteExpiredVersionsCommand{}
	if err := bus.Dispatch(&cmd); err != nil {
		srv.log.Error("无法删除过期的仪表板版本", "error", err.Error())
	} else {
		srv.log.Debug("删除旧/过期的仪表板版本", "rows affected", cmd.DeletedRows)
	}
}

func (srv *CleanUpService) deleteOldLoginAttempts() {
	if srv.Cfg.DisableBruteForceLoginProtection {
		return
	}

	cmd := m.DeleteOldLoginAttemptsCommand{
		OlderThan: time.Now().Add(time.Minute * -10),
	}
	if err := bus.Dispatch(&cmd); err != nil {
		srv.log.Error("删除过期登录尝试时出现问题", "error", err.Error())
	} else {
		srv.log.Debug("删除过期的登录尝试次数", "rows affected", cmd.DeletedRows)
	}
}
