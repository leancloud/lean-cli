package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func msgWithRegion(msg string) string {
	// If failed to detect the current region, just return the error message as is.
	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return msg
	}
	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return msg
	}
	return fmt.Sprintf("User doesn't sign in at region %s.", region)
}

func wrapAction(action cli.ActionFunc) cli.ActionFunc {
	prefix := color.RedString("[ERROR]")
	return func(c *cli.Context) error {
		err := action(c)
		switch e := err.(type) {
		case nil:
			return nil
		case api.Error:
			var msg string
			// Make error message more friendly to users having applications at multi regions.
			if strings.HasPrefix(e.Content, "unauthorized") {
				msg = msgWithRegion("User doesn't sign in.")
			} else {
				msg = e.Content
			}
			return cli.NewExitError(fmt.Sprintf("%s %s", prefix, msg), 1)
		case *cli.ExitError:
			return e
		default:
			return cli.NewExitError(fmt.Sprintf("%s %s", prefix, err.Error()), 1)
		}
	}
}
