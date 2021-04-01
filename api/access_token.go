package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
)

type accessTokenMapping map[string]regions.Region

var accessTokenCache accessTokenMapping

func init() {
	accessTokenCache = make(map[string]regions.Region)
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

func getAccessTokenRegion(region regions.Region) string {
	for k, v := range accessTokenCache {
		if v == region {
			return k
		}
	}

	return ""
}

func (cache accessTokenMapping) Add(accessKey string, region regions.Region) accessTokenMapping {
	for k, v := range cache {
		if v == region {
			delete(cache, k)
		}
	}
	cache[accessKey] = region
	return cache
}

func (cache accessTokenMapping) Remove(accessKey string) accessTokenMapping {
	delete(cache, accessKey)
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
