package main

import (
	"fmt"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/runtimes"
)

// newCliError create a *cli.ExitError from given error interface, and the new error's content is human readable
func newCliError(err error) *cli.ExitError {
	switch err {
	case runtimes.ErrInvalidRuntime:
		msg := fmt.Sprintf(" %s 错误的项目目录结构，请确保当前运行目录是正确的云引擎项目", chrysanthemum.Fail)
		return cli.NewExitError(msg, 1)
	}

	switch e := err.(type) {
	case api.Error:
		return cli.NewExitError(fmt.Sprintf(" %s %s", chrysanthemum.Fail, e.Content), 1)
	case *cli.ExitError:
		return e
	default:
		return cli.NewExitError(fmt.Sprintf(" %s %s", chrysanthemum.Fail, err.Error()), 1)
	}
}
