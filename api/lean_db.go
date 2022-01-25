package api

import (
	"fmt"
	"strconv"
	"strings"
)

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
