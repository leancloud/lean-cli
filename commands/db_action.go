package commands

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"

	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/proxy"
	"github.com/urfave/cli"
)

func getLeanDBClusterList(appID string) (api.LeanDBClusterSlice, error) {
	clusters, err := api.GetLeanDBClusterList(appID)
	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, cli.NewExitError("This app doesn't have any LeanDB instance", 1)
	}

	sort.Sort(sort.Reverse(clusters))

	return clusters, nil
}

func dbListAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	clusters, err := getLeanDBClusterList(appID)
	if err != nil {
		return err
	}

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

func selectDbCluster(clusters []*api.LeanDBCluster) (*api.LeanDBCluster, error) {
	var selectedCluster *api.LeanDBCluster
	question := wizard.Question{
		Content: "Please choose a LeanDB instance",
		Answers: []wizard.Answer{},
	}
	m := make(map[string]bool)
	for _, cluster := range clusters {
		runtimeName := fmt.Sprintf("%s-%s", cluster.Runtime, cluster.Name)
		var content string
		if ok := m[runtimeName]; ok {
			content = fmt.Sprintf("%s (shared from %s) - %s", cluster.Name, cluster.AppID, cluster.NodeQuota)
		} else {
			content = fmt.Sprintf("%s - %s", cluster.Name, cluster.NodeQuota)
		}
		m[runtimeName] = true

		answer := wizard.Answer{
			Content: content,
		}
		// for scope problem
		func(cluster *api.LeanDBCluster) {
			answer.Handler = func() {
				selectedCluster = cluster
			}
		}(cluster)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedCluster, err
}

func parseProxyInfo(c *cli.Context) (*proxy.ProxyInfo, error) {
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return nil, err
	}
	clusters, err := getLeanDBClusterList(appID)
	if err != nil {
		return nil, err
	}

	var cluster *api.LeanDBCluster
	localPort := c.Int("port")
	instanceName := c.Args().Get(0)

	if instanceName == "" {
		instance, err := selectDbCluster(clusters)
		if err != nil {
			return nil, err
		}
		cluster = instance
	} else {
		proxyAppID := c.String("app-id")
		if proxyAppID == "" {
			proxyAppID = appID
		}
		for _, c := range clusters {
			if c.Name == instanceName && c.AppID == proxyAppID {
				cluster = c
			}
		}
		if cluster == nil {
			s := fmt.Sprintf("No instance for [%s (%s)]", instanceName, proxyAppID)
			return nil, cli.NewExitError(s, 1)
		}
	}

	if cluster.Status != "running" && cluster.Status != "updating" && cluster.Status != "recovering" {
		s := fmt.Sprintf("instance [%s] is in [%s] status, not one of accessible status [running, updating, recovering]", cluster.Name, cluster.Status)
		return nil, cli.NewExitError(s, 1)
	}

	p := &proxy.ProxyInfo{
		AppID:        cluster.AppID,
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

func dbShellAction(c *cli.Context) error {
	p, err := parseProxyInfo(c)
	if err != nil {
		return err
	}
	if clis := proxy.RuntimeClis[p.Runtime]; clis == nil {
		s := fmt.Sprintf("LeanDB runtime %s don't support shell proxy.", p.Runtime)
		return cli.NewExitError(s, 1)
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
