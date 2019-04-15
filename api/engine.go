package api

import (
	"errors"
	"net/url"

	"github.com/levigross/grequests"
)

type EngineInfo struct {
	AppID string `json:"appId"`
}

type VersionInfo struct {
	VersionTag string `json:"versionTag"`
}

type GroupDeployInfo struct {
	Deployable bool        `json:"deployable"`
	Version    VersionInfo `json:"version"`
}

type InstanceInfo struct {
	Name string `json:"name"`
	Prod int    `json:"prod"`
}

type GetGroupsResult struct {
	GroupName    string            `json:"groupName"`
	Repository   string            `json:"repository"`
	Domain       string            `json:"domain"`
	Instances    []InstanceInfo    `json:"instances"`
	Staging      GroupDeployInfo   `json:"staging"`
	Environments map[string]string `json:"environments"`
}

type DeployOptions struct {
	Message     string
	NoDepsCache bool
	Options     string // Additional options in urlencode format
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
		url = "/1.1/engine/groups/" + group + "/staging/version"
	case 1:
		url = "/1.1/engine/groups/" + group + "/production/version"
	default:
		return nil, errors.New("invalid prod value " + string(prod))
	}

	return client.post(url, params, opts)
}

// DeployImage will deploy the engine group with specify image tag
func DeployImage(appID string, group string, prod int, imageTag string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["versionTag"] = imageTag

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
func DeployAppFromGit(appID string, group string, prod int, revision string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["gitTag"] = revision

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
func DeployAppFromFile(appID string, group string, prod int, fileURL string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["zipUrl"] = fileURL

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

// GetGroups returns the application's engine groups
func GetGroups(appID string) ([]*GetGroupsResult, error) {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	resp, err := client.get("/1.1/engine/groups?all=true", opts)
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
	return nil, errors.New("Failed to find group: " + groupName)
}

func GetEngineInfo(appID string) (*EngineInfo, error) {
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
	var result = new(EngineInfo)
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
		return errors.New("Error updating environment variable, code: " + string(response.StatusCode))
	}
	return nil
}

func prepareDeployParams(options *DeployOptions) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"noDependenciesCache": options.NoDepsCache,
		"async":               true,
	}

	if options.Message != "" {
		params["comment"] = options.Message
	}

	if options.Options != "" {
		queryString, err := url.ParseQuery(options.Options)

		if err != nil {
			return nil, err
		}

		for k, v := range queryString {
			params[k] = v[0]
		}
	}

	return params, nil
}
