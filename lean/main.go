package main

import (
	"github.com/codegangsta/cli"
	"os"
)

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
		println(subCmd)
	}

	app.Run(os.Args)
}
