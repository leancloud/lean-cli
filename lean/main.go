package main

import (
	"github.com/codegangsta/cli"
	"os/exec"
	"os"
	"fmt"
)

func callSubCmd(cmd string) error {
	return exec.Command(cmd).Run()
}

func main() {
	app := cli.NewApp()

	app.Name = "lean"
	app.Usage = "make an explosive entrance"
	app.Version = VERSION

	app.Flags = []cli.Flag{}

	app.Action = func(ctx *cli.Context) {
		if len(ctx.Args()) == 0 {
			cli.ShowAppHelp(ctx)
			return
		}
		subCmd := ctx.Args()[0]
		subCmd = "lean-" + subCmd
		err := callSubCmd(subCmd)
		fmt.Fprintln(os.Stderr, "Call sub command error:")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app.Run(os.Args)
}
