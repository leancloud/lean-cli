package api

import (
	"fmt"

	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

type RouterResponse struct {
	TTL             int    `json:"ttl"`
	StatsServer     string `json:"stats_server"`
	RTMRouterServer string `json:"rtm_router_server"`
	PushServer      string `json:"push_server"`
	EngineServer    string `json:"engine_server"`
	APIServer       string `json:"api_server"`
}

// Not applicable for US
func QueryAppRouter(appID string) (result RouterResponse, err error) {
	resp, err := grequests.Get("https://app-router.leancloud.cn/2/route?appId="+appID, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return result, err
	}
	if !resp.Ok {
		return result, fmt.Errorf("query app router failed: %d", resp.StatusCode)
	}

	if err = resp.JSON(&result); err != nil {
		return result, err
	}

	return result, nil
}
