package api

import (
	"errors"

	"github.com/leancloud/lean-cli/lean/api/regions"
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
	client := NewClient(region)

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
		routerCache[app.AppID] = region
	}
	if err = saveRouterCache(); err != nil {
		return nil, err
	}

	return result, nil
}

// DeployImage will deploy the engine group with specify image tag
func DeployImage(appID string, groupName string, imageTag string) (string, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return "", err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return "", err
	}
	opts.Headers["X-LC-Id"] = appID

	resp, err := client.put("/1.1/engine/groups/"+groupName+"/deploy", map[string]interface{}{
		"imageTag": imageTag,
		"async":    true,
	}, opts)

	if err != nil {
		return "", err
	}
	result := new(struct {
		EventToken string `json:"eventToken"`
	})

	err = resp.JSON(result)
	return result.EventToken, err
}

// DeployAppFromGit will deploy applications with user's git repo
// returns the event token for polling deploy log
func DeployAppFromGit(appID string, projectPath string, groupName string, revision string, noDepsCache bool) (string, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return "", err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return "", err
	}
	opts.Headers["X-LC-Id"] = appID

	resp, err := client.post("/1.1/engine/groups/"+groupName+"/buildAndDeploy", map[string]interface{}{
		"comment":             "",
		"noDependenciesCache": noDepsCache,
		"async":               true,
		"gitTag":              revision,
	}, opts)

	if err != nil {
		return "", err
	}

	result := new(struct {
		EventToken string `json:"eventToken"`
	})
	err = resp.JSON(result)
	return result.EventToken, err
}

// DeployAppFromFile will deploy applications with specific file
// returns the event token for polling deploy log
func DeployAppFromFile(appID string, projectPath string, groupName string, fileURL string, message string, noDepsCache bool) (string, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return "", err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return "", err
	}
	opts.Headers["X-LC-Id"] = appID

	resp, err := client.post("/1.1/engine/groups/"+groupName+"/buildAndDeploy", map[string]interface{}{
		"zipUrl":              fileURL,
		"comment":             message,
		"noDependenciesCache": noDepsCache,
		"async":               true,
	}, opts)

	if err != nil {
		return "", err
	}

	result := new(struct {
		EventToken string `json:"eventToken"`
	})
	err = resp.JSON(result)
	return result.EventToken, err
}

// GetAppInfoResult is GetAppInfo function's result type
type GetAppInfoResult struct {
	AppDomain      string `json:"app_domain"`
	AppID          string `json:"app_id"`
	AppKey         string `json:"app_key"`
	AppName        string `json:"app_name"`
	HookKey        string `json:"hook_key"`
	LeanEngineMode string `json:"leanengine_mode"`
	MasterKey      string `json:"master_key"`
}

// GetAppInfo returns the application's detail info
func GetAppInfo(appID string) (*GetAppInfoResult, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return nil, err
	}
	client := NewClient(region)

	resp, err := client.get("/1.1/clients/self/apps/"+appID, nil)
	if err != nil {
		return nil, err
	}
	result := new(GetAppInfoResult)
	err = resp.JSON(result)
	return result, err
}

// GetGroupsResult is GetGroups's result struct
type GetGroupsResult struct {
	GroupName string `json:"groupName"`
	Prod      int    `json:"prod"`
	Instances []struct {
		Name  string `json:"name"`
		Quota int    `json:"quota"`
	} `json:"instances"`
	CurrentImage struct {
		Runtime  string `json:"runtime"`
		ImageTag string `json:"imageTag"`
	} `json:"currentImage"`
}

// GetGroups returns the application's engine groups
func GetGroups(appID string) ([]*GetGroupsResult, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return nil, err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	resp, err := client.get("/1.1/engine/groups", opts)
	if err != nil {
		return nil, err
	}

	var result []*GetGroupsResult
	err = resp.JSON(&result)

	return result, err
}

type GetEngineInfoResult struct {
	AppID         string            `json:"appId"`
	Mode          string            `json:"mode"`
	InstanceLimit int               `json:"instanceLimit"`
	Version       string            `json:"version"`
	Environments  map[string]string `json:"environments"`
}

func GetEngineInfo(appID string) (*GetEngineInfoResult, error) {
	region, err := GetAppRegion(appID)
	if err != nil {
		return nil, err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	response, err := client.get("/1.1/functions/_ops/engine", opts)
	if err != nil {
		return nil, err
	}
	var result = new(GetEngineInfoResult)
	err = response.JSON(result)
	return result, err
}

func PutEnvironments(appID string, envs map[string]string) error {
	region, err := GetAppRegion(appID)
	if err != nil {
		return err
	}
	client := NewClient(region)

	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appID

	params := make(map[string]interface{})
	for k, v := range envs {
		params[k] = v
	}

	response, err := client.put("/1.1/functions/_ops/engine/environments", params, opts)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("update environment failed")
	}
	return nil
}
