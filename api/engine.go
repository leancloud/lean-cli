package api

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/levigross/grequests"
	"github.com/mattn/go-colorable"
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
	Production   GroupDeployInfo   `json:"production"`
	Environments map[string]string `json:"environments"`
}

type DeployOptions struct {
	DirectUpload   bool
	Message        string
	NoDepsCache    bool
	OverwriteFuncs bool
	BuildLogs      bool
	Commit         string
	Url            string
	Options        string // Additional options in urlencode format
}

func deploy(appID string, group string, env string, params map[string]interface{}) (*grequests.Response, error) {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID

	url := fmt.Sprintf("/1.1/engine/groups/%s/envs/%s/version", group, env)

	directUpload, _ := params["direct"].(bool)
	delete(params, "direct")
	if directUpload {
		opts.Data = func() map[string]string {
			data := make(map[string]string)
			for k, v := range params {
				data[k] = fmt.Sprint(v)
			}
			return data
		}()
		archiveFilePath := opts.Data["zipUrl"]
		delete(opts.Data, "zipUrl")
		fd, err := os.Open(archiveFilePath)
		if err != nil {
			return nil, err
		}

		stats, err := fd.Stat()
		if err != nil {
			return nil, err
		}

		bar := pb.New(int(stats.Size())).SetUnits(pb.U_BYTES).SetMaxWidth(80)
		bar.Output = colorable.NewColorableStderr()
		bar.Prefix(color.GreenString("[INFO]") + " Uploading file")
		bar.Start()
		barProxy := bar.NewProxyReader(fd)

		opts.Files = []grequests.FileUpload{
			{
				FileName:     "leanengine.zip",
				FileContents: barProxy,
				FieldName:    "zip",
			},
		}

		resp, err := client.post(url, nil, opts)
		if err != nil {
			return nil, err
		}
		bar.Finish()

		return resp, nil
	}
	return client.post(url, params, opts)
}

// DeployImage will deploy the engine group with specify image tag
func DeployImage(appID string, group string, env string, imageTag string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["versionTag"] = imageTag

	resp, err := deploy(appID, group, env, params)
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
func DeployAppFromGit(appID string, group string, env string, revision string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["gitTag"] = revision

	resp, err := deploy(appID, group, env, params)
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
func DeployAppFromFile(appID string, group string, env string, fileURL string, opts *DeployOptions) (string, error) {
	params, err := prepareDeployParams(opts)

	if err != nil {
		return "", err
	}

	params["direct"] = opts.DirectUpload
	params["zipUrl"] = fileURL

	resp, err := deploy(appID, group, env, params)
	if err != nil {
		return "", err
	}

	result := new(struct {
		EventToken string `json:"eventToken"`
	})
	err = resp.JSON(result)
	return result.EventToken, err
}

func DeleteEnvironment(appID string, group string, env string) error {
	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appID

	_, err = client.delete(fmt.Sprintf("/1.1/engine/groups/%s/envs/%s", group, env), opts)
	return err
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
		return fmt.Errorf("Error updating environment variable, code: %d", response.StatusCode)
	}
	return nil
}

func prepareDeployParams(options *DeployOptions) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"noDependenciesCache": options.NoDepsCache,
		"overwriteFunctions":  options.OverwriteFuncs,
		"async":               true,
		"printBuildLogs":      options.BuildLogs,
	}

	if options.Message != "" {
		params["comment"] = options.Message
	}
	if options.Commit != "" {
		params["commit"] = options.Commit
	}
	if options.Url != "" {
		params["url"] = options.Url
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
