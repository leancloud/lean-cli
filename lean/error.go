package main

import (
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/runtimes"
)

// newCliError create a *cli.ExitError from given error interface, and the new error's content is human readable
func newCliError(err error) *cli.ExitError {
	switch err {
	case runtimes.ErrInvalidRuntime:
		msg := color.RedString("> 错误的项目目录结构，请确保当前运行目录是正确的云引擎项目")
		return cli.NewExitError(msg, 1)
	}

	switch e := err.(type) {
	case api.Error:
		return cli.NewExitError(color.RedString("> %s", e.Content), 1)
	case *cli.ExitError:
		return e
	default:
		return cli.NewExitError(color.RedString("> %s", err.Error()), 1)
	}
}
