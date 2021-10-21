package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

var defaultAPIURL = map[regions.Region]string{
	regions.ChinaNorth: "https://api.leancloud.cn",
	regions.USWest:     "https://us-api.leancloud.cn",
	regions.ChinaEast:  "https://tab.leancloud.cn",
}

type RouterResponse struct {
	TTL             int    `json:"ttl"`
	StatsServer     string `json:"stats_server"`
	RTMRouterServer string `json:"rtm_router_server"`
	PushServer      string `json:"push_server"`
	EngineServer    string `json:"engine_server"`
	APIServer       string `json:"api_server"`
}

func GetAppAPIURL(region regions.Region, appID string) string {
	envAPIURL := os.Getenv("LEANCLOUD_API_SERVER")

	if envAPIURL != "" {
		return envAPIURL
	}

	switch region {
	case regions.ChinaNorth, regions.ChinaEast:
		routerInfo, err := QueryAppRouter(appID)

		if err != nil {
			logp.Warn(err) // Ignore app router error
		} else {
			return "https://" + routerInfo.APIServer
		}
	case regions.ChinaTDS1:
		return "https://" + strings.ToLower(appID[0:8]) + ".cloud.tds1.tapapis.cn"
	case regions.APSG:
		return "https://" + strings.ToLower(appID[0:8]) + ".cloud.ap-sg.tapapis.com"
	}

	return defaultAPIURL[region]
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
