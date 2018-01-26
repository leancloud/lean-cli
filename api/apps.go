package api

import (
	"errors"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/levigross/grequests"
)

// GetAppListResult is GetAppList function's result type
type GetAppListResult struct {
	AppID     string `json:"app_id"`
	AppKey    string `json:"app_key"`
	AppName   string `json:"app_name"`
	MasterKey string `json:"master_key"`
	AppDomain string `json:"app_domain"`
}

const (
	DEPLOY_SMOOTHLY = "smoothly"
	DEPLOY_ATOMIC   = "atomic"
)

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
		routerCache[app.AppID] = region
	}
	if err = saveRouterCache(); err != nil {
		return nil, err
	}

	return result, nil
}

func deploy(appID string, group string, prod int, params map[string]interface{}) (*grequests.Response, error) {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	var url string
	switch prod {
	case 0:
		url = "/1.1/engine/groups/" + group + "/stagingImage"
	case 1:
		url = "/1.1/engine/groups/" + group + "/productionImage"
	default:
		return nil, errors.New("invalid prod value " + string(prod))
	}

	return client.post(url, params, opts)
}

// DeployImage will deploy the engine group with specify image tag
func DeployImage(appID string, group string, prod int, imageTag string, mode string) (string, error) {
	params := map[string]interface{}{
		"imageTag": imageTag,
		"async":    true,
	}

	params[mode] = true

	resp, err := deploy(appID, group, prod, params)
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
func DeployAppFromGit(appID string, group string, prod int, revision string, noDepsCache bool, mode string) (string, error) {
	params := map[string]interface{}{
		"noDependenciesCache": noDepsCache,
		"async":               true,
		"gitTag":              revision,
	}

	params[mode] = true

	resp, err := deploy(appID, group, prod, params)
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
func DeployAppFromFile(appID string, group string, prod int, fileURL string, message string, noDepsCache bool, mode string) (string, error) {
	params := map[string]interface{}{
		"zipUrl":              fileURL,
		"comment":             message,
		"noDependenciesCache": noDepsCache,
		"async":               true,
	}

	params[mode] = true

	resp, err := deploy(appID, group, prod, params)
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

// GetGroupsResult is GetGroups's result struct
type GetGroupsResult struct {
	GroupName  string `json:"groupName"`
	Repository string `json:"repository"`
	Domain     string `json:"domain"`
	Instances  []struct {
		Name  string `json:"name"`
		Quota int    `json:"quota"`
	} `json:"instances"`
	StagingImage struct {
		Runtime  string `json:"runtime"`
		ImageTag string `json:"imageTag"`
	} `json:"stagingImage"`
	Environments map[string]string `json:"environments"`
}

// GetGroups returns the application's engine groups
func GetGroups(appID string) ([]*GetGroupsResult, error) {
	client := NewClientByApp(appID)

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
	if err != nil {
		return nil, err
	}

	// filter the staging group, since it's not used anymore
	var filtered []*GetGroupsResult
	for _, group := range result {
		if group.GroupName == "staging" {
			continue
		}
		filtered = append(filtered, group)
	}

	return filtered, nil
}

// GetGroup will fetch all groups from API and return the current group info
func GetGroup(appID string, groupName string) (*GetGroupsResult, error) {
	groups, err := GetGroups(appID)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.GroupName == groupName {
			return group, nil
		}
	}
	return nil, errors.New("找不到分组：" + groupName)
}

type GetEngineInfoResult struct {
	AppID         string            `json:"appId"`
	Mode          string            `json:"mode"`
	InstanceLimit int               `json:"instanceLimit"`
	Version       string            `json:"version"`
	Environments  map[string]string `json:"environments"`
}

func GetEngineInfo(appID string) (*GetEngineInfoResult, error) {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	response, err := client.get("/1.1/engine", opts)
	if err != nil {
		return nil, err
	}
	var result = new(GetEngineInfoResult)
	err = response.JSON(result)
	return result, err
}

func PutEnvironments(appID string, group string, envs map[string]string) error {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appID

	params := make(map[string]interface{})
	environments := make(map[string]interface{})
	for k, v := range envs {
		environments[k] = v
	}
	params["environments"] = environments

	url := "/1.1/engine/groups/" + group
	response, err := client.patch(url, params, opts)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("更新运引擎环境变量失败，响应码：" + string(response.StatusCode))
	}
	return nil
}
