package api

import (
	"fmt"

	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/levigross/grequests"
)

var routerCache = make(map[string]regions.Region)

// GetAppRegion will query the app router, and return the app's region.
// The result is cached in process memory.
func GetAppRegion(appID string) (regions.Region, error) {
	if r, ok := routerCache[appID]; ok {
		return r, nil
	}
	resp, err := grequests.Get("https://app-router.leancloud.cn/1/route?appId="+appID, nil)
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
	default:
		return regions.Invalid, fmt.Errorf("invalid region server: %s", result.APIServer)
	}
}
