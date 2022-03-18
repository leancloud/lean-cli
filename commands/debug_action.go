package commands

import (
	"strconv"

	"github.com/aisk/logp"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/console"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func debugAction(c *cli.Context) error {
	version.PrintVersionAndEnvironment()
	remote := c.String("remote")
	port := strconv.Itoa(c.Int("port"))
	appID := c.String("app-id")

	if appID == "" {
		var err error
		appID, err = apps.GetCurrentAppID(".")
		if err != nil {
			return err
		}
	}

	logp.Info("Retrieving app info ...")
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}
	logp.Infof("Current app: %s (%s)\r\n", color.RedString(appInfo.AppName), appID)

	cons := &console.Server{
		AppID:       appInfo.AppID,
		AppKey:      appInfo.AppKey,
		MasterKey:   appInfo.MasterKey,
		HookKey:     appInfo.HookKey,
		RemoteURL:   remote,
		ConsolePort: port,
		Errors:      make(chan error),
	}

	cons.Run()
	for {
		select {
		case err = <-cons.Errors:
			panic(err)
		}
	}
}
