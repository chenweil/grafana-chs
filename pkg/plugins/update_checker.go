package plugins

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/hashicorp/go-version"
)

var (
	httpClient = http.Client{Timeout: 10 * time.Second}
)

type GrafanaNetPlugin struct {
	Slug    string `json:"slug"`
	Version string `json:"version"`
}

type GithubLatest struct {
	Stable  string `json:"stable"`
	Testing string `json:"testing"`
}

func getAllExternalPluginSlugs() string {
	var result []string
	for _, plug := range Plugins {
		if plug.IsCorePlugin {
			continue
		}

		result = append(result, plug.Id)
	}

	return strings.Join(result, ",")
}

func (pm *PluginManager) checkForUpdates() {
	if !setting.CheckForUpdates {
		return
	}

	pm.log.Debug("Checking for updates")

	pluginSlugs := getAllExternalPluginSlugs()
	resp, err := httpClient.Get("https://grafana.com/api/plugins/versioncheck?slugIn=" + pluginSlugs + "&grafanaVersion=" + setting.BuildVersion)

	if err != nil {
		log.Trace("无法从网站获得插件回购, %v", err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Trace("更新检查失败，从网站读取响应, %v", err.Error())
		return
	}

	gNetPlugins := []GrafanaNetPlugin{}
	err = json.Unmarshal(body, &gNetPlugins)
	if err != nil {
		log.Trace("无法解组插件回购，阅读网站的回复, %v", err.Error())
		return
	}

	for _, plug := range Plugins {
		for _, gplug := range gNetPlugins {
			if gplug.Slug == plug.Id {
				plug.GrafanaNetVersion = gplug.Version

				plugVersion, err1 := version.NewVersion(plug.Info.Version)
				gplugVersion, err2 := version.NewVersion(gplug.Version)

				if err1 != nil || err2 != nil {
					plug.GrafanaNetHasUpdate = plug.Info.Version != plug.GrafanaNetVersion
				} else {
					plug.GrafanaNetHasUpdate = plugVersion.LessThan(gplugVersion)
				}
			}
		}
	}

	resp2, err := httpClient.Get("https://raw.githubusercontent.com/grafana/grafana/master/latest.json")
	if err != nil {
		log.Trace("无法从github.com获得latest.json repo: %v", err.Error())
		return
	}

	defer resp2.Body.Close()
	body, err = ioutil.ReadAll(resp2.Body)
	if err != nil {
		log.Trace("更新检查失败，从github.com读取响应, %v", err.Error())
		return
	}

	var githubLatest GithubLatest
	err = json.Unmarshal(body, &githubLatest)
	if err != nil {
		log.Trace("无法解组github.com最新，阅读来自github.com的回复: %v", err.Error())
		return
	}

	if strings.Contains(setting.BuildVersion, "-") {
		GrafanaLatestVersion = githubLatest.Testing
		GrafanaHasUpdate = !strings.HasPrefix(setting.BuildVersion, githubLatest.Testing)
	} else {
		GrafanaLatestVersion = githubLatest.Stable
		GrafanaHasUpdate = githubLatest.Stable != setting.BuildVersion
	}

	currVersion, err1 := version.NewVersion(setting.BuildVersion)
	latestVersion, err2 := version.NewVersion(GrafanaLatestVersion)

	if err1 == nil && err2 == nil {
		GrafanaHasUpdate = currVersion.LessThan(latestVersion)
	}
}
