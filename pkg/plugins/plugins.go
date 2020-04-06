package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util"
)

var (
	DataSources  map[string]*DataSourcePlugin
	Panels       map[string]*PanelPlugin
	StaticRoutes []*PluginStaticRoute
	Apps         map[string]*AppPlugin
	Plugins      map[string]*PluginBase
	PluginTypes  map[string]interface{}
	Renderer     *RendererPlugin

	GrafanaLatestVersion string
	GrafanaHasUpdate     bool
	plog                 log.Logger
)

type PluginScanner struct {
	pluginPath string
	errors     []error
}

type PluginManager struct {
	log log.Logger
}

func init() {
	registry.RegisterService(&PluginManager{})
}

func (pm *PluginManager) Init() error {
	pm.log = log.New("plugins")
	plog = log.New("plugins")

	DataSources = map[string]*DataSourcePlugin{}
	StaticRoutes = []*PluginStaticRoute{}
	Panels = map[string]*PanelPlugin{}
	Apps = map[string]*AppPlugin{}
	Plugins = map[string]*PluginBase{}
	PluginTypes = map[string]interface{}{
		"panel":      PanelPlugin{},
		"datasource": DataSourcePlugin{},
		"app":        AppPlugin{},
		"renderer":   RendererPlugin{},
	}

	pm.log.Info("Starting plugin search")
	scan(path.Join(setting.StaticRootPath, "app/plugins"))

	// check if plugins dir exists
	if _, err := os.Stat(setting.PluginsPath); os.IsNotExist(err) {
		if err = os.MkdirAll(setting.PluginsPath, os.ModePerm); err != nil {
			plog.Error("无法创建插件目录", "dir", setting.PluginsPath, "error", err)
		} else {
			plog.Info("插件目录创建", "dir", setting.PluginsPath)
			scan(setting.PluginsPath)
		}
	} else {
		scan(setting.PluginsPath)
	}

	// check plugin paths defined in config
	checkPluginPaths()

	for _, panel := range Panels {
		panel.initFrontendPlugin()
	}

	for _, ds := range DataSources {
		ds.initFrontendPlugin()
	}

	for _, app := range Apps {
		app.initApp()
	}

	return nil
}

func (pm *PluginManager) startBackendPlugins(ctx context.Context) error {
	for _, ds := range DataSources {
		if ds.Backend {
			if err := ds.startBackendPlugin(ctx, plog); err != nil {
				pm.log.Error("无法初始化插件。", "error", err, "plugin", ds.Id)
			}
		}
	}

	return nil
}

func (pm *PluginManager) Run(ctx context.Context) error {
	pm.startBackendPlugins(ctx)
	pm.updateAppDashboards()
	pm.checkForUpdates()

	ticker := time.NewTicker(time.Minute * 10)
	run := true

	for run {
		select {
		case <-ticker.C:
			pm.checkForUpdates()
		case <-ctx.Done():
			run = false
		}
	}

	// kill backend plugins
	for _, p := range DataSources {
		p.Kill()
	}

	return ctx.Err()
}

func checkPluginPaths() error {
	for _, section := range setting.Raw.Sections() {
		if strings.HasPrefix(section.Name(), "plugin.") {
			path := section.Key("path").String()
			if path != "" {
				scan(path)
			}
		}
	}
	return nil
}

func scan(pluginDir string) error {
	scanner := &PluginScanner{
		pluginPath: pluginDir,
	}

	if err := util.Walk(pluginDir, true, true, scanner.walker); err != nil {
		if pluginDir != "data/plugins" {
			log.Warn("无法扫描目录 \"%v\" 错误: %s", pluginDir, err)
		}
		return err
	}

	if len(scanner.errors) > 0 {
		return errors.New("某些插件无法加载")
	}

	return nil
}

func (scanner *PluginScanner) walker(currentPath string, f os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if f.Name() == "node_modules" {
		return util.ErrWalkSkipDir
	}

	if f.IsDir() {
		return nil
	}

	if f.Name() == "plugin.json" {
		err := scanner.loadPluginJson(currentPath)
		if err != nil {
			log.Error(3, "插件：无法加载插件json文件: %v,  err: %v", currentPath, err)
			scanner.errors = append(scanner.errors, err)
		}
	}
	return nil
}

func (scanner *PluginScanner) loadPluginJson(pluginJsonFilePath string) error {
	currentDir := filepath.Dir(pluginJsonFilePath)
	reader, err := os.Open(pluginJsonFilePath)
	if err != nil {
		return err
	}

	defer reader.Close()

	jsonParser := json.NewDecoder(reader)
	pluginCommon := PluginBase{}
	if err := jsonParser.Decode(&pluginCommon); err != nil {
		return err
	}

	if pluginCommon.Id == "" || pluginCommon.Type == "" {
		return errors.New("在plugin.json中找不到type和id属性")
	}

	var loader PluginLoader
	pluginGoType, exists := PluginTypes[pluginCommon.Type]
	if !exists {
		return errors.New("未知的插件类型 " + pluginCommon.Type)
	}
	loader = reflect.New(reflect.TypeOf(pluginGoType)).Interface().(PluginLoader)

	reader.Seek(0, 0)
	return loader.Load(jsonParser, currentDir)
}

func GetPluginMarkdown(pluginId string, name string) ([]byte, error) {
	plug, exists := Plugins[pluginId]
	if !exists {
		return nil, PluginNotFoundError{pluginId}
	}

	path := filepath.Join(plug.PluginDir, fmt.Sprintf("%s.md", strings.ToUpper(name)))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join(plug.PluginDir, fmt.Sprintf("%s.md", strings.ToLower(name)))
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make([]byte, 0), nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}
