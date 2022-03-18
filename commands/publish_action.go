package commands

import (
	"errors"
	"fmt"

	"github.com/aisk/logp"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func publishAction(c *cli.Context) error {
	version.PrintVersionAndEnvironment()
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("Please use `lean checkout` to designate a LeanCloud app first.", 1)
	}
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}
	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		return err
	}

	if !groupInfo.Staging.Deployable {
		return errors.New("staging environment not available for trial version")
	}

	logp.Info(fmt.Sprintf("Current app: %s (%s), group: %s, region: %s", color.GreenString(appInfo.AppName), appID, color.GreenString(groupName), region))
	logp.Info(fmt.Sprintf("Deploying verison %s to %s", groupInfo.Staging.Version.VersionTag, color.GreenString("production")))

	tok, err := api.DeployImage(appID, groupName, 1, groupInfo.Staging.Version.VersionTag, &api.DeployOptions{
		OverwriteFuncs: c.Bool("overwrite-functions"),
		Options:        c.String("options"),
	})

	if err != nil {
		return err
	}

	ok, err := api.PollEvents(appID, tok)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Deployment failed")
	}
	return nil
}
