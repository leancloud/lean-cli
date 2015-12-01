package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/codegangsta/cli"
)

var commands = []cli.Command{
	{
		Name:   "create",
		Usage:  "Create a new LeanEngine app",
		Action: subCmdCliHandler("lean-create"),
	},
	{
		Name:   "init",
		Usage:  "Initialize current folder as a LeanEngine app",
		Action: subCmdCliHandler("lean-init"),
	},
	{
		Name:   "run",
		Usage:  "Run the LeanEngine app in local machine",
		Action: subCmdCliHandler("lean-run"),
	},
	{
		Name:   "deploy",
		Usage:  "Deploy the app to LeanEngine",
		Action: subCmdCliHandler("lean-deploy"),
	},
}

func callSubCmd(cmd string) error {
	return exec.Command(cmd).Run()
}

func subCmdCliHandler(cmd string) func(*cli.Context) {
	return func(ctx *cli.Context) {
		if err := callSubCmd("lean-create"); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Call sub command error:")
			os.Exit(1)
		}
	}
}

func main() {
	app := cli.NewApp()

	app.Name = path.Base(os.Args[0])
	app.Usage = "LeanCloud command line tool"
	app.Version = VERSION

	app.Flags = []cli.Flag{}

	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
