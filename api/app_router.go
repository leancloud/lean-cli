package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

var routerCache = make(map[string]regions.Region)

// GetAppRegion will query the app router, and return the app's region.
// The result is cached in process memory.
func GetAppRegion(appID string) (regions.Region, error) {
	if r, ok := routerCache[appID]; ok {
		return r, nil
	}
	resp, err := grequests.Get("https://app-router.leancloud.cn/1/route?appId="+appID, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return regions.Invalid, err
	}
	if !resp.Ok {
		return regions.Invalid, fmt.Errorf("query app router failed: %d", resp.StatusCode)
	}

	var result struct {
		APIServer string `json:"api_server"`
	}
	if err = resp.JSON(&result); err != nil {
		return regions.Invalid, err
	}

	switch result.APIServer {
	case "us-api.leancloud.cn":
		routerCache[appID] = regions.US
		return regions.US, nil
	case "api.leancloud.cn":
		routerCache[appID] = regions.CN
		return regions.CN, nil
	case "e1-api.leancloud.cn":
		routerCache[appID] = regions.TAB
		return regions.TAB, nil
	default:
		return regions.Invalid, fmt.Errorf("invalid region server: %s", result.APIServer)
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
