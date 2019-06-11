package api

import (
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
)

// GetAppListResult is GetAppList function's result type
type GetAppListResult struct {
	AppID     string `json:"app_id"`
	AppKey    string `json:"app_key"`
	AppName   string `json:"app_name"`
	MasterKey string `json:"master_key"`
	AppDomain string `json:"app_domain"`
}

// GetAppList returns the current user's all LeanCloud application
// this will also update the app router cache
func GetAppList(region regions.Region) ([]*GetAppListResult, error) {
	client := NewClientByRegion(region)

	resp, err := client.get("/1/clients/self/apps", nil)
	if err != nil {
		return nil, err
	}

	var result []*GetAppListResult
	err = resp.JSON(&result)
	if err != nil {
		return nil, err
	}

	for _, app := range result {
		apps.SetRegionCache(app.AppID, region)
	}

	if err = apps.SaveRegionCache(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetAppInfoResult is GetAppInfo function's result type
type GetAppInfoResult struct {
	AppDomain string `json:"app_domain"`
	AppID     string `json:"app_id"`
	AppKey    string `json:"app_key"`
	AppName   string `json:"app_name"`
	HookKey   string `json:"hook_key"`
	MasterKey string `json:"master_key"`
}

// GetAppInfo returns the application's detail info
func GetAppInfo(appID string) (*GetAppInfoResult, error) {
	client := NewClientByApp(appID)

	resp, err := client.get("/1.1/clients/self/apps/"+appID, nil)
	if err != nil {
		return nil, err
	}
	result := new(GetAppInfoResult)
	err = resp.JSON(result)
	return result, err
}
