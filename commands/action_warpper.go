package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/urfave/cli"
)

func wrapAction(action cli.ActionFunc) cli.ActionFunc {
	prefix := color.RedString("[ERROR]")
	return func(c *cli.Context) error {
		err := action(c)
		switch e := err.(type) {
		case nil:
			return nil
		case api.Error:
			return cli.NewExitError(fmt.Sprintf("%s %s", prefix, e.Content), 1)
		case *cli.ExitError:
			return e
		default:
			return cli.NewExitError(fmt.Sprintf("%s %s", prefix, err.Error()), 1)
		}
	}
}
