package api

import (
	"github.com/bitly/go-simplejson"
)

// AppDetail returns the app's detail infomation
func (client *Client) AppDetail() (*simplejson.Json, error) {
	return client.get("/__leancloud/apps/appDetail", nil)
}

// EngineInfo returns the app's engine infomation
func (client *Client) EngineInfo() (*simplejson.Json, error) {
	return client.get("/functions/_ops/engine", nil)
}

// Groups ...
func (client *Client) Groups() (*simplejson.Json, error) {
	return client.get("/functions/_ops/groups", nil)
}

// BuildFromURL ...
func (client *Client) BuildFromURL(groupName string, fileURL string) (*simplejson.Json, error) {
	return client.post("/functions/_ops/groups/"+groupName+"/buildAndDeploy", map[string]interface{}{
		"zipUrl":              fileURL,
		"comment":             "",
		"noDependenciesCache": false,
		"async":               true,
	}, nil)
}

// BuildFromGit ...
func (client *Client) BuildFromGit(groupName string) (*simplejson.Json, error) {
	return client.post("/functions/_ops/groups/"+groupName+"/buildAndDeploy", map[string]interface{}{
		"comment":             "",
		"noDependenciesCache": false,
		"async":               true,
	}, nil)
}
