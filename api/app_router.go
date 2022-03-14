package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/leancloud/lean-cli/api/regions"
)

// URL for LeanStorage API, not dashboard API
func GetAppAPIURL(region regions.Region, appID string) string {
	envAPIURL := os.Getenv("LEANCLOUD_API_SERVER")

	if envAPIURL != "" {
		return envAPIURL
	}

	shortAppId := strings.ToLower(appID[0:8])

	switch region {
	case regions.ChinaNorth:
		return fmt.Sprint("https://", shortAppId, ".lc-cn-n1-shared.com")
	case regions.USWest:
		return fmt.Sprint("https://", shortAppId, ".api.lncldglobal.com")
	case regions.ChinaEast:
		return fmt.Sprint("https://", shortAppId, ".lc-cn-e1-shared.com")
	case regions.ChinaTDS1:
		return fmt.Sprint("https://", shortAppId, ".cloud.tds1.tapapis.cn")
	case regions.APSG:
		return fmt.Sprint("https://", shortAppId, ".cloud.ap-sg.tapapis.com")
	default:
		panic(fmt.Errorf("invalid region: %s", region))
	}
}
