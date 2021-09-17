package commands

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/proxy"
	"github.com/urfave/cli"
)

func dbListAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	// TODO filter not running instances
	clusters, err := api.GetLeanDBClusterList(appID)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	sort.Sort(clusters)

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	fmt.Fprintln(t, "InstanceName\t\t\tQuota")
	for _, cluster := range clusters {
		// TODO show appId
		fmt.Fprintf(t, "%s\t\t\t%s\r\n", cluster.Name, cluster.NodeQuota)
	}
	t.Flush()

	return nil
}

func parseProxyInfo(c *cli.Context) (*proxy.ProxyInfo, error) {
	if c.NArg() < 1 {
		// TODO show subcommand help
		cli.ShowCommandHelp(c, "")
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return nil, err
	}

	clusters, err := api.GetLeanDBClusterList(appID)
	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	localPort := c.Int("port")
	proxyAppID := c.String("app-id")
	if proxyAppID == "" {
		proxyAppID = appID
	}

	instanceName := c.Args().Get(0)
	var cluster *api.LeanDBCluster
	for _, c := range clusters {
		if c.Name == instanceName && c.Status == "running" && c.AppID == proxyAppID {
			cluster = c
		}
	}

	if cluster == nil {
		s := fmt.Sprintf("No running instance for %s", instanceName)
		return nil, cli.NewExitError(s, 1)
	}

	proxyInfo := &proxy.ProxyInfo{
		AppID:        proxyAppID,
		ClusterId:    cluster.ID,
		Name:         cluster.Name,
		Runtime:      cluster.Runtime,
		AuthUser:     cluster.AuthUser,
		AuthPassword: cluster.AuthPassword,
		LocalPort:    strconv.Itoa(localPort),
		Connected:    make(chan bool, 1),
	}

	return proxyInfo, nil
}

func dbProxyAction(c *cli.Context) error {
	proxyInfo, err := parseProxyInfo(c)
	if err != nil {
		return err
	}

	logp.Infof("Proxy to LeanDB instance %s(%s) on local port %s\r\n", proxyInfo.Name, proxyInfo.AppID, proxyInfo.LocalPort)

	return proxy.Run(proxyInfo)
}

func dbShellAction(c *cli.Context) error {
	proxyInfo, err := parseProxyInfo(c)
	if err != nil {
		return err
	}

	go func() {
		<-proxyInfo.Connected
		proxy.ForkExecCli(proxyInfo)
	}()

	return proxy.Run(proxyInfo)
}
