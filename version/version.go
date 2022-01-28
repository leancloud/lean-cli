package version

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api/regions"
)

// Version is lean-cli's version.
const Version = "0.29.2"

var Distribution string

var defaultRegionMapping = map[string]regions.Region{
	"lean": regions.ChinaNorth,
	"tds":  regions.ChinaTDS1,
}

var availableRegionsMapping = map[string][]regions.Region{
	"lean": {regions.ChinaNorth, regions.USWest, regions.ChinaEast},
	"tds":  {regions.ChinaTDS1, regions.APSG},
}

var LoginViaAccessTokenOnly bool
var DefaultRegion regions.Region
var AvailableRegions []regions.Region

func init() {
	Distribution = filepath.Base(os.Args[0])
	Distribution = strings.TrimSuffix(Distribution, filepath.Ext(Distribution))
	if idx := strings.Index(Distribution, "-"); idx != -1 {
		Distribution = Distribution[:idx]
	}
	if Distribution != "lean" && Distribution != "tds" {
		logp.Warnf("Invalid executable name: `%s`, falling back to `lean`.\n", Distribution)
		logp.Warn("Please rename the executable to `lean` or `tds` depending on whether your app is on LeanCloud or TDS.")
		Distribution = "lean"
	}

	LoginViaAccessTokenOnly = Distribution == "tds"
	DefaultRegion = defaultRegionMapping[Distribution]
	AvailableRegions = availableRegionsMapping[Distribution]
}

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
