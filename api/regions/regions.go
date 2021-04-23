package regions

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/utils"
)

// Region is region's type
type Region int

func Parse(region string) Region {
	switch region {
	case "cn", "CN", "cn-n1":
		return ChinaNorth
	case "tab", "TAB", "cn-e1":
		return ChinaEast
	case "us", "US", "us-w1":
		return USWest
	default:
		return Invalid
	}
}

var regionLoginStatus = make(map[Region]bool)

func init() {
	regionStatus, err := ioutil.ReadFile(filepath.Join(utils.ConfigDir(), "leancloud", "logined_regions.json"))
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		if err := json.Unmarshal(regionStatus, &regionLoginStatus); err != nil {
			panic(err)
		}
	}
}
func (r Region) String() string {
	switch r {
	case ChinaNorth:
		return "cn-n1"
	case ChinaEast:
		return "cn-e1"
	case USWest:
		return "us-w1"
	default:
		return "invalid"
	}
}

func (r Region) EnvString() string {
	switch r {
	case ChinaNorth, ChinaEast:
		return "CN"
	case USWest:
		return "US"
	default:
		return "invalid"
	}
}

// Description is region's readable description
func (r Region) Description() string {
	switch r {
	case ChinaNorth:
		return "China North"
	case USWest:
		return "United States"
	case ChinaEast:
		return "China East"
	default:
		return "invalid"
	}
}

// API server regions
const (
	Invalid Region = iota
	ChinaNorth
	USWest
	ChinaEast
)

func GetLoginedRegions() []Region {
	var regions []Region
	for k, v := range regionLoginStatus {
		if v {
			regions = append(regions, k)
		}
	}

	return regions
}

func GetRegionLoginStatus() map[Region]bool {
	return regionLoginStatus
}

func SetRegionLoginStatus(region Region) {
	regionLoginStatus[region] = true
}

func SaveRegionLoginStatus() error {
	data, err := json.MarshalIndent(regionLoginStatus, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(utils.ConfigDir(), "leancloud", "logined_regions.json"), data, 0644)
}
