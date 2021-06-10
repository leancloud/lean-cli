package version

import (
	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api/regions"
)

// Version is lean-cli's version.
const Version = "0.25.0"

var Distribution = "lean"

var LoginViaAccessTokenOnly = Distribution == "tds"

var defaultRegionMapping = map[string]regions.Region{
	"lean": regions.ChinaNorth,
	"tds":  regions.ChinaTDS1,
}

var availableRegionsMapping = map[string][]regions.Region{
	"lean": {regions.ChinaNorth, regions.USWest, regions.ChinaEast},
	"tds":  {regions.ChinaTDS1},
}

var DefaultRegion = defaultRegionMapping[Distribution]
var AvailableRegions = availableRegionsMapping[Distribution]

func PrintCurrentVersion() {
	logp.Info("Current CLI tool version: ", Version)
}

func GetUserAgent() string {
	switch Distribution {
	case "lean":
		return "LeanCloud-CLI/" + Version
	case "tds":
		return "TDS-CLI/" + Version
	}

	panic("invalid distribution")
}
