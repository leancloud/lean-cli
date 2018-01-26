package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
)

var routerCache = make(map[string]regions.Region)

func GetAppRegion(appID string) (regions.Region, error) {
	if r, ok := routerCache[appID]; ok {
		return r, nil
	} else {
		return regions.Invalid, errors.New("应用配置信息不完整，请重新运行 `lean switch` 关联应用")
	}
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
