package remotecache

import (
	"testing"

	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
)

// NewFakeStore creates store for testing
func NewFakeStore(t *testing.T) *RemoteCache {
	t.Helper()

	opts := &setting.RemoteCacheOptions{
		Name:    "database",
		ConnStr: "",
	}

	SQLStore := sqlstore.InitTestDB(t)

	dc := &RemoteCache{
		SQLStore: SQLStore,
		Cfg: &setting.Cfg{
			RemoteCacheOptions: opts,
		},
	}

	err := dc.Init()
	if err != nil {
		t.Fatalf("无法初始化远程缓存进行测试。 错误: %v", err)
	}

	return dc
}
