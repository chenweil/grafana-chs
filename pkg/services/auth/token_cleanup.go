package auth

import (
	"context"
	"time"

	"github.com/grafana/grafana/pkg/services/sqlstore"
)

func (srv *UserAuthTokenService) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Hour)
	maxInactiveLifetime := time.Duration(srv.Cfg.LoginMaxInactiveLifetimeDays) * 24 * time.Hour
	maxLifetime := time.Duration(srv.Cfg.LoginMaxLifetimeDays) * 24 * time.Hour

	err := srv.ServerLockService.LockAndExecute(ctx, "cleanup expired auth tokens", time.Hour*12, func() {
		srv.deleteExpiredTokens(ctx, maxInactiveLifetime, maxLifetime)
	})

	if err != nil {
		srv.log.Error("无法锁定并执行已过期的身份验证令牌的清理", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			err := srv.ServerLockService.LockAndExecute(ctx, "cleanup expired auth tokens", time.Hour*12, func() {
				srv.deleteExpiredTokens(ctx, maxInactiveLifetime, maxLifetime)
			})

			if err != nil {
				srv.log.Error("无法锁定并执行已过期的身份验证令牌的清理", "error", err)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *UserAuthTokenService) deleteExpiredTokens(ctx context.Context, maxInactiveLifetime, maxLifetime time.Duration) (int64, error) {
	createdBefore := getTime().Add(-maxLifetime)
	rotatedBefore := getTime().Add(-maxInactiveLifetime)

	srv.log.Debug("开始清理过期的身份验证令牌", "createdBefore", createdBefore, "rotatedBefore", rotatedBefore)

	var affected int64
	err := srv.SQLStore.WithDbSession(ctx, func(dbSession *sqlstore.DBSession) error {
		sql := `DELETE from user_auth_token WHERE created_at <= ? OR rotated_at <= ?`
		res, err := dbSession.Exec(sql, createdBefore.Unix(), rotatedBefore.Unix())
		if err != nil {
			return err
		}

		affected, err = res.RowsAffected()
		if err != nil {
			srv.log.Error("无法清除过期的身份验证令牌", "error", err)
			return nil
		}

		srv.log.Debug("完成过期的auth令牌的清理", "count", affected)

		return nil
	})

	return affected, err
}
