package api

import "fmt"

// ExecuteCacheCommandResult is ExecuteClusterCommand's result type
type ExecuteCacheCommandResult struct {
	Result interface{} `json:"result"`
}

// LeanCacheCluster is structure of LeanCache DB instannce
type LeanCacheCluster struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Runtime   string `json:"runtime"`
	NodeQuota string `json:"nodeQuota"`
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
