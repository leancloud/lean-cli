package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/urfave/cli"
)

// newCliError create a *cli.ExitError from given error interface, and the new error's content is human readable
func newCliError(err error) *cli.ExitError {
	prefix := color.RedString("[ERROR]")
	switch err {
	case runtimes.ErrInvalidRuntime:
		msg := fmt.Sprintf("%s 错误的项目目录结构，请确保当前运行目录是正确的云引擎项目", prefix)
		return cli.NewExitError(msg, 1)
	}

	switch e := err.(type) {
	case api.Error:
		return cli.NewExitError(fmt.Sprintf("%s %s", prefix, e.Content), 1)
	case *cli.ExitError:
		return e
	default:
		return cli.NewExitError(fmt.Sprintf("%s %s", prefix, err.Error()), 1)
	}
}
