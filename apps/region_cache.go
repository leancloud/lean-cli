package apps

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
)

var ErrMissingRegionCache = errors.New("App configuration is incomplete. Please run `lean switch` to configure the app.")

var regionCache = make(map[string]regions.Region)

func GetAppRegion(appID string) (regions.Region, error) {
	if r, ok := regionCache[appID]; ok {
		return r, nil
	} else {
		return regions.Invalid, ErrMissingRegionCache
	}
}

func GetLoginedRegions() (result []regions.Region) {
	for _, region := range regionCache {
		if !regionInArray(region, result) {
			result = append(result, region)
		}
	}

	return result
}

func SetRegionCache(appID string, region regions.Region) {
	regionCache[appID] = region
}

func SaveRegionCache() error {
	data, err := json.MarshalIndent(regionCache, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(utils.ConfigDir(), "leancloud", "app_router.json"), data, 0644)
}

func init() {
	data, err := ioutil.ReadFile(filepath.Join(utils.ConfigDir(), "leancloud", "app_router.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(data, &regionCache); err != nil {
			panic(err)
		}
	}
}

func regionInArray(region regions.Region, list []regions.Region) bool {
	for _, r := range list {
		if r == region {
			return true
		}
	}
	return false
}
