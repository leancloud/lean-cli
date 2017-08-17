package commands

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func thirdPartyCommand(c *cli.Context, _cmdName string) {
	cmdName := "lean-" + _cmdName

	// executeble not found:

	execPath, err := exec.LookPath(filepath.Join(".leancloud", "bin", cmdName))

	if err != nil {
		execPath, err = exec.LookPath(cmdName)
		if e, ok := err.(*exec.Error); ok {
			if e.Err == exec.ErrNotFound {
				cli.ShowAppHelp(c)
				return
			}
			log.Fatal(err)
		} else if err != nil {
			log.Fatal(err)
		}
	}

	cmd := exec.Command(execPath, c.Args()[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	appID, err := apps.GetCurrentAppID(".")
	if err == nil {
		region, err := api.GetAppRegion(appID)
		if err != nil {
			log.Fatal(err)
		}
		appInfo, err := api.GetAppInfo(appID)
		if err != nil {
			log.Fatal(err)
		}
		envs := []string{
			"LEANCLOUD_APP_ID=" + appInfo.AppID,
			"LEANCLOUD_APP_KEY=" + appInfo.AppKey,
			"LEANCLOUD_APP_MASTER_KEY=" + appInfo.MasterKey,
			"LEANCLOUD_APP_HOOK_KEY=" + appInfo.HookKey,
			"LEANCLOUD_APP_ENV=" + "development",
			"LEANCLOUD_REGION=" + region.String(),
		}
		for _, env := range envs {
			cmd.Env = append(cmd.Env, env)
		}
	}
	if err != nil && err != apps.ErrNoAppLinked {
		log.Fatal(err)
	}

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
