package api

import (
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
)

// GetAppListResult is GetAppList function's result type
type GetAppListResult struct {
	AppID     string `json:"appId"`
	AppKey    string `json:"appKey"`
	AppName   string `json:"appName"`
	MasterKey string `json:"masterKey"`
	AppDomain string `json:"appDomain"`
}

// GetAppList returns the current user's all LeanCloud application
// this will also update the app router cache
func GetAppList(region regions.Region) ([]*GetAppListResult, error) {
	client := NewClientByRegion(region)

	resp, err := client.get("/client-center/2/clients/self/apps", nil)
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
	AppDomain string `json:"appDomain"`
	AppID     string `json:"appId"`
	AppKey    string `json:"appKey"`
	AppName   string `json:"appName"`
	HookKey   string `json:"hookKey"`
	MasterKey string `json:"masterKey"`
}

// GetAppInfo returns the application's detail info
func GetAppInfo(appID string) (*GetAppInfoResult, error) {
	client := NewClientByApp(appID)

	resp, err := client.get("/client-center/2/clients/self/apps/"+appID, nil)
	if err != nil {
		return nil, err
	}
	result := new(GetAppInfoResult)
	err = resp.JSON(result)
	return result, err
}
