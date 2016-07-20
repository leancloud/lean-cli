package main

import (
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
)

// newCliError create a *cli.ExitError from given error interface, and the new error's content is human readable
func newCliError(err error) *cli.ExitError {
	switch e := err.(type) {
	case api.Error:
		return cli.NewExitError(e.Content, 1)
	default:
		return cli.NewExitError(err.Error(), 1)
	}
}
