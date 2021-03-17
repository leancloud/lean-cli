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

	return runClusterListAction(appID)
}

func runClusterListAction(appID string) error {
	clusters, err := api.GetClusterList(appID)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	fmt.Fprintln(t, "InstanceName\tQuota")
	for _, cluster := range clusters {
		fmt.Fprintf(t, "%s\t%s\r\n", cluster.Name, cluster.NodeQuota)
	}
	t.Flush()

	return nil
}
