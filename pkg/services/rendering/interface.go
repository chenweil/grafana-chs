package rendering

import (
	"context"
	"errors"
	"time"

	"github.com/grafana/grafana/pkg/models"
)

var ErrTimeout = errors.New("超时错误。 您可以使用＆timeout url参数设置超时（以秒为单位）")
var ErrNoRenderer = errors.New("找不到渲染器插件，也没有配置外部渲染服务器")
var ErrPhantomJSNotInstalled = errors.New("找不到PhantomJS可执行文件")

type Opts struct {
	Width           int
	Height          int
	Timeout         time.Duration
	OrgId           int64
	UserId          int64
	OrgRole         models.RoleType
	Path            string
	Encoding        string
	Timezone        string
	ConcurrentLimit int
}

type RenderResult struct {
	FilePath string
}

type renderFunc func(ctx context.Context, options Opts) (*RenderResult, error)

type Service interface {
	Render(ctx context.Context, opts Opts) (*RenderResult, error)
}
