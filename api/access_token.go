package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
)

type accessTokenMapping map[regions.Region]string

var accessTokenCache accessTokenMapping

func init() {
	accessTokenCache = make(map[regions.Region]string)
	content, err := ioutil.ReadFile(filepath.Join(utils.ConfigDir(), "leancloud", "access-tokens"))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(content, &(accessTokenCache)); err != nil {
			panic(err)
		}
	}
}

func getAccessTokenByRegion(region regions.Region) string {
	for k, v := range accessTokenCache {
		if k == region {
			return v
		}
	}

	return ""
}

func (cache accessTokenMapping) Add(accessToken string, region regions.Region) accessTokenMapping {
	cache[region] = accessToken
	return cache
}

func (cache accessTokenMapping) Save() error {
	keysCache, err := json.Marshal(accessTokenCache)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(utils.ConfigDir(), "leancloud", "access-tokens"), keysCache, 0600); err != nil {
		return err
	}

	return nil
}
