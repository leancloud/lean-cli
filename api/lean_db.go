package api

import (
	"fmt"
	"strconv"
	"strings"
)

// ExecuteCacheCommandResult is ExecuteClusterCommand's result type
type ExecuteCacheCommandResult struct {
	Result interface{} `json:"result"`
}

// LeanCacheCluster is structure of LeanCache DB instannce
// TODO remove when remove `lean cache` in 1.0
type LeanCacheCluster struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Runtime   string `json:"runtime"`
	NodeQuota string `json:"nodeQuota"`
}

// {
// "id": 415,
// "appId": "hIksfdh8IFpkeDsoceTR5Hwu-gzGzoHsz",
// "name": "yqiu_test_es",
// "runtime": "es",
// "nodeQuota": "es-512",
// "storageQuota": "1H",
// "dataNodes": 1,
// "status": "running",
// "proxyPort": 27214,
// "authUser": "elasticsearch",
// "authPassword": "g08pBQ0XknasKy1w",
// "createdAt": "2021-08-10T07:56:01.000Z",
// "updatedAt": "2021-08-10T08:00:50.000Z",
// "version": 3,
// "versionTag": "7.9.2",
// "proxyHost": "engine-stateful-proxy.engine.svc.cn-n1"
// }

type LeanDBCluster struct {
	ID           int    `json:"id"`
	AppID        string `json:"appId"`
	Name         string `json:"name"`
	Runtime      string `json:"runtime"`
	NodeQuota    string `json:"nodeQuota"`
	Status       string `json:"status"`
	AuthUser     string `json:"authUser"`
	AuthPassword string `json:"authPassword"`
}

type LeanDBClusterSlice []*LeanDBCluster

func (x LeanDBClusterSlice) Len() int {
	return len(x)
}

// NodeQuota: es-512 es-1024 mongo-512 redis-128 udb-500
// compare: runtime -> quota -> id
func (x LeanDBClusterSlice) Less(i, j int) bool {
	l, r := x[i], x[j]
	if l.Runtime == r.Runtime {
		lm, _ := strconv.Atoi(strings.Split(l.NodeQuota, "-")[1])
		rm, _ := strconv.Atoi(strings.Split(r.NodeQuota, "-")[1])
		if lm == rm {
			return l.ID < r.ID
		}
		return lm < rm
	}
	return l.Runtime < r.Runtime
}

func (x LeanDBClusterSlice) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func GetLeanDBClusterList(appID string) (LeanDBClusterSlice, error) {
	client := NewClientByApp(appID)

	url := fmt.Sprintf("/1.1/leandb/apps/%s/clusters", appID)
	resp, err := client.get(url, nil)
	if err != nil {
		return nil, err
	}

	var result LeanDBClusterSlice
	err = resp.JSON(&result)

	if err != nil {
		return nil, err
	}

	return result, err
}

// GetClusterList returns current app's LeanCache instances (NEW)
func GetClusterList(appID string) ([]*LeanCacheCluster, error) {
	client := NewClientByApp(appID)

	url := fmt.Sprintf("/1.1/leandb/apps/%s/clusters", appID)
	resp, err := client.get(url, nil)
	if err != nil {
		return nil, err
	}

	var result []*LeanCacheCluster
	err = resp.JSON(&result)

	if err != nil {
		return nil, err
	}

	return result, err
}

// ExecuteClusterCommand will send command to LeanCache and excute it
func ExecuteClusterCommand(appID string, clusterID int, db int, command string) (*ExecuteCacheCommandResult, error) {
	client := NewClientByApp(appID)

	url := fmt.Sprintf("/1.1/leandb/clusters/%d/user-command/exec", clusterID)
	resp, err := client.post(url, map[string]interface{}{
		"db":      db,
		"command": command}, nil)

	if err != nil {
		return nil, err
	}

	result := new(ExecuteCacheCommandResult)
	err = resp.JSON(result)

	return result, err
}
