package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func cacheListAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	caches, err := api.GetCacheList(appID)
	if err != nil {
		return newCliError(err)
	}

	if len(caches) == 0 {
		return cli.NewExitError("该应用没有 LeanCache 实例", 1)
	}

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	fmt.Fprintln(t, "InstanceName\tMaxMemory")
	for _, cache := range caches {
		fmt.Fprintf(t, "%s\t%dM\r\n", cache.Instance, cache.MaxMemory)
	}
	t.Flush()

	return nil
}
