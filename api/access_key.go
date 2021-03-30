package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
)

type accessKey struct {
	Keys  map[string]regions.Region `json:"keys"`
	saved bool
}

var accessKeyCache accessKey

func init() {
	accessKeyCache.Keys = make(map[string]regions.Region)
	content, err := ioutil.ReadFile(filepath.Join(utils.ConfigDir(), "leancloud", "keys"))
	if err != nil {
		if os.IsNotExist(err) {
			if file, err := os.OpenFile(filepath.Join(utils.ConfigDir(), "leancloud", "keys"), os.O_CREATE, 0644); err != nil {
				panic(err)
			} else {
				file.WriteString("{}")
				if err := file.Close(); err != nil {
					panic(err)
				}
			}
		} else {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(content, &(accessKeyCache.Keys)); err != nil {
			panic(err)
		}
	}
}

func (key *accessKey) Add(accessKey string, region regions.Region) *accessKey {
	for k, v := range key.Keys {
		if v == region {
			delete(key.Keys, k)
		}
	}
	key.Keys[accessKey] = region
	return key
}

func (key *accessKey) Remove(accessKey string) *accessKey {
	delete(key.Keys, accessKey)
	return key
}

func (key *accessKey) Save() error {
	keysCache, err := json.Marshal(accessKeyCache.Keys)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(utils.ConfigDir(), "leancloud", "keys"), keysCache, 0664); err != nil {
		return err
	}

	return nil
}
