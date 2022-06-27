package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api/regions"
)

// Version is lean-cli's version.
const Version = "1.0.1"

var Distribution string

var defaultRegionMapping = map[string]regions.Region{
	"lean": regions.ChinaNorth,
	"tds":  regions.ChinaTDS1,
}

var availableRegionsMapping = map[string][]regions.Region{
	"lean": {regions.ChinaNorth, regions.USWest, regions.ChinaEast},
	"tds":  {regions.ChinaTDS1, regions.APSG},
}

var brandNameMapping = map[string]string{
	"lean": "LeanCloud",
	"tds":  "TapTap Developer Services",
}

var engineBrandNameMapping = map[string]string{
	"lean": "LeanEngine",
	"tds":  "Cloud Engine",
}

var dbBrandNameMapping = map[string]string{
	"lean": "LeanDB",
	"tds":  "Database",
}

var LoginViaAccessTokenOnly bool
var DefaultRegion regions.Region
var AvailableRegions []regions.Region

var BrandName string
var EngineBrandName string
var DBBrandName string

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

	BrandName = brandNameMapping[Distribution]
	EngineBrandName = engineBrandNameMapping[Distribution]
	DBBrandName = dbBrandNameMapping[Distribution]
}

func PrintVersionAndEnvironment() {
	// Print all environment info to improve the efficiency of technical support
	logp.Info(fmt.Sprintf("%s (v%s) running on %s/%s", os.Args[0], Version, runtime.GOOS, runtime.GOARCH))
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
