package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

const banner = `
 _                        ______ _                 _
| |                      / _____) |               | |
| |      ____ ____ ____ | /     | | ___  _   _  _ | |
| |     / _  ) _  |  _ \| |     | |/ _ \| | | |/ || |
| |____( (/ ( ( | | | | | \_____| | |_| | |_| ( (_| |
|_______)____)_||_|_| |_|\______)_|\___/ \____|\____|

`

const version = "0.0.1"

func thirdPartyCommand(c *cli.Context, _cmd string) {
	cmd := "lean-" + _cmd
	println(cmd)
}

func main() {
	// disable the log prefix
	log.SetFlags(0)

	// add banner text to help text
	cli.AppHelpTemplate = banner + cli.AppHelpTemplate

	app := cli.NewApp()
	app.Name = "lean"
	app.Version = version
	app.Usage = "Command line to manage and deploy LeanCloud apps"

	app.CommandNotFound = thirdPartyCommand

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
		{
			Name:  "app",
			Usage: "多应用管理，可以使用一个云引擎项目关联多个 LeanCloud 应用",
			Subcommands: []cli.Command{
				{
					Name: "list",
				},
				{
					Name: "add",
				},
				{
					Name: "switch",
				},
				{
					Name: "remove",
				},
			},
		},
		{
			Name:   "deploy",
			Usage:  "部署云引擎项目到服务器",
			Action: deployAction,
		},
	}

	app.Run(os.Args)
}
