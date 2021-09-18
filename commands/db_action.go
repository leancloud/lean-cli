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

	clusters, err := api.GetLeanDBClusterList(appID)
	if err != nil {
		return err
	}

	if len(clusters) == 0 {
		return cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	sort.Sort(sort.Reverse(clusters))

	t := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	m := make(map[string]bool)
	fmt.Fprintln(t, "InstanceName\t\t\tQuota")
	for _, cluster := range clusters {
		runtimeName := fmt.Sprintf("%s-%s", cluster.Runtime, cluster.Name)
		if ok := m[runtimeName]; ok {
			fmt.Fprintf(t, "%s (shared from %s)\t\t\t%s\r\n", cluster.Name, cluster.AppID, cluster.NodeQuota)
		} else {
			fmt.Fprintf(t, "%s\t\t\t%s\r\n", cluster.Name, cluster.NodeQuota)
		}
		m[runtimeName] = true
	}
	t.Flush()

	return nil
}

func parseProxyInfo(c *cli.Context) (*proxy.ProxyInfo, error) {
	if c.NArg() < 1 {
		cli.ShowSubcommandHelp(c)
		return nil, cli.NewExitError("", 1)
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
		if c.Name == instanceName && c.AppID == proxyAppID {
			cluster = c
		}
	}

	if cluster == nil {
		s := fmt.Sprintf("No instance for [%s (%s)]", instanceName, proxyAppID)
		return nil, cli.NewExitError(s, 1)
	} else if cluster.Status != "running" {
		s := fmt.Sprintf("instance [%s] is not in running status", instanceName)
		return nil, cli.NewExitError(s, 1)
	}

	p := &proxy.ProxyInfo{
		AppID:        proxyAppID,
		ClusterId:    cluster.ID,
		Name:         cluster.Name,
		Runtime:      cluster.Runtime,
		AuthUser:     cluster.AuthUser,
		AuthPassword: cluster.AuthPassword,
		LocalPort:    strconv.Itoa(localPort),
	}

	return p, nil
}

func dbProxyAction(c *cli.Context) error {
	p, err := parseProxyInfo(c)
	if err != nil {
		return err
	}

	return proxy.RunProxy(p)
}

var runtimeShell = map[string]bool{
	"redis": true,
	"udb":   true,
	"mysql": true,
	"mongo": true,
}

func dbShellAction(c *cli.Context) error {
	p, err := parseProxyInfo(c)
	if err != nil {
		return err
	}
	if ok := runtimeShell[p.Runtime]; !ok {
		msg := fmt.Sprintf("LeanDB runtime %s don't support shell proxy.", p.Runtime)
		return cli.NewExitError(msg, 1)
	}

	started := make(chan bool, 1)
	term := make(chan bool, 1)
	go func() {
		<-started
		err := proxy.ForkExecCli(p, term)
		if err != nil {
			logp.Warnf("Start cli get error: %s", err)
			term <- true
		}
	}()

	return proxy.RunShellProxy(p, started, term)
}
