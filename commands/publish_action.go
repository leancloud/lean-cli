package commands

import (
	"errors"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func publishAction(c *cli.Context) error {
	version.PrintCurrentVersion()
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

	logp.Info("Retrieving app info ...")
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
		return errors.New("For development apps, `lean deploy` directly deploys to production. There is no need to use this command.")
	}

	logp.Infof("Deploying %s(%s) to region: %s group: %s production\r\n", appInfo.AppName, appID, region, groupName)

	tok, err := api.DeployImage(appID, groupName, 1, groupInfo.Staging.Version.VersionTag, &api.DeployOptions{
		Options: c.String("options"),
	})

	ok, err := api.PollEvents(appID, tok)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("Deployment failed")
	}
	return nil
}
