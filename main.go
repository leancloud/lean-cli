package main

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
)

func main() {
	// disable the log prefix
	log.SetFlags(0)

	app := cli.NewApp()
	app.Name = "lean"

	app.Commands = []cli.Command{
		{
			Name:  "up",
			Usage: "本地启动云引擎应用。",
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:   "new",
			Usage:  "创建云引擎项目。",
			Action: newAction,
		},
	}

	app.Run(os.Args)
}
