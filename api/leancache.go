package api

import "fmt"

// GetCacheListResult is GetCacheList's return structure type
type GetCacheListResult struct {
	Instance   string `json:"instance"`
	MaxMemory  int    `json:"max_memory"`
	InstanceID string `json:"instance_id"`
	Info       struct {
		UsedMemoryHuman string `json:"used_memory_human"`
	} `json:"info"`
}

// GetCacheList returns current app's LeanCache instance list
func GetCacheList(appID string) ([]*GetCacheListResult, error) {
	// TODO: we will use cookie based auth latter
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return nil, err
	}

	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID
	opts.Headers["X-LC-Key"] = appInfo.MasterKey + ",master"

	resp, err := client.get("/1.1/__cache/ops/instances", opts)
	if err != nil {
		return nil, err
	}

	var result []*GetCacheListResult
	err = resp.JSON(&result)

	return result, err
}

// ExecuteCacheCommandResult is ExecuteCacheCommand's result type
type ExecuteCacheCommandResult struct {
	Result interface{} `json:"result"`
}

// ExecuteCacheCommand will send command to LeanCache and excute it
func ExecuteCacheCommand(appID string, instance string, db int, command string) (*ExecuteCacheCommandResult, error) {
	// TODO: we will use cookie based auth latter
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return nil, err
	}

	client := NewClientByApp(appID)

	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appID
	opts.Headers["X-LC-Key"] = appInfo.MasterKey + ",master"

	url := fmt.Sprintf("/1.1/__cache/ops/instances/%s/dbs/%d", instance, db)

	resp, err := client.post(url, map[string]interface{}{
		"command": command,
	}, opts)

	if err != nil {
		return nil, err
	}

	result := new(ExecuteCacheCommandResult)
	err = resp.JSON(result)

	return result, err
}
