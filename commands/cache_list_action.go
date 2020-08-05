package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func cacheListAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	ver, err := api.GetVersion(appID)
	if err != nil {
		return err
	}

	fmt.Printf("lean cache version: %d\n", ver)

	switch ver {
	case 0:
		return runCacheListAction(appID)
	case 1:
		return runInstanceListAction(appID)
	default:
		return cli.NewExitError("The app cannot use lean cache list.", 1)
	}
}

func runCacheListAction(appID string) error {
	caches, err := api.GetCacheList(appID)
	if err != nil {
		return err
	}

	if len(caches) == 0 {
		return cli.NewExitError("This app doesn't have any LeanCache instance", 1)
	}

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	fmt.Fprintln(t, "InstanceName\tMaxMemory")
	for _, cache := range caches {
		fmt.Fprintf(t, "%s\t%dM\r\n", cache.Instance, cache.MaxMemory)
	}
	t.Flush()

	return nil
}

func runInstanceListAction(appID string) error {
	instances, err := api.GetClusterList(appID)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		return cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	fmt.Fprintln(t, "InstanceName\tQuota")
	for _, instance := range instances {
		fmt.Fprintf(t, "%s\t%s\r\n", instance.Name, instance.NodeQuota)
	}
	t.Flush()

	return nil
}
