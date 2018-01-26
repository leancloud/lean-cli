package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/coreos/etcd/version"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
	"github.com/levigross/grequests"
)

var routerCache = make(map[string]regions.Region)

type RouterResponse struct {
	TTL             int    `json:"ttl"`
	StatsServer     string `json:"stats_server"`
	RTMRouterServer string `json:"rtm_router_server"`
	PushServer      string `json:"push_server"`
	EngineServer    string `json:"engine_server"`
	APIServer       string `json:"api_server"`
}

func GetAppRegion(appID string) (regions.Region, error) {
	if r, ok := routerCache[appID]; ok {
		return r, nil
	} else {
		return regions.Invalid, errors.New("应用配置信息不完整，请重新运行 `lean switch` 关联应用")
	}
}

// Not applicable for US
func QueryAppRouter(appID string) (result RouterResponse, err error) {
	resp, err := grequests.Get("https://app-router.leancloud.cn/2/route?appId="+appID, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return result, err
	}
	if !resp.Ok {
		return result, fmt.Errorf("query app router failed: %d", resp.StatusCode)
	}

	if err = resp.JSON(&result); err != nil {
		return result, err
	}

	return result, nil
}

func saveRouterCache() error {
	data, err := json.MarshalIndent(routerCache, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(utils.ConfigDir(), "leancloud", "app_router.json"), data, 0644)
}

func init() {
	data, err := ioutil.ReadFile(filepath.Join(utils.ConfigDir(), "leancloud", "app_router.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &routerCache)
}
