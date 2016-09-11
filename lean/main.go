package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/logo"
	"github.com/leancloud/lean-cli/lean/output"
	"github.com/leancloud/lean-cli/lean/version"
)

var (
	isDeployFromGit = false
	op              = output.NewOutput(os.Stdout)
)

func thirdPartyCommand(c *cli.Context, _cmdName string) {
	cmdName := "lean-" + _cmdName

	// executeble not found:
	execPath, err := exec.LookPath(cmdName)
	if e, ok := err.(*exec.Error); ok {
		if e.Err == exec.ErrNotFound {
			cli.ShowAppHelp(c)
			return
		}
	}
	cmd := exec.Command(execPath, c.Args()[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// disable the log prefix
	log.SetFlags(0)

	go func() {
		_ = checkUpdate()
	}()

	// add banner text to help text
	cli.AppHelpTemplate = logo.Logo() + cli.AppHelpTemplate
	cli.SubcommandHelpTemplate = logo.Logo() + cli.SubcommandHelpTemplate

	app := cli.NewApp()
	app.Name = "lean"
	app.Version = version.Version
	app.Usage = "Command line to manage and deploy LeanCloud apps"
	app.EnableBashCompletion = true

	app.CommandNotFound = thirdPartyCommand

	app.Commands = []cli.Command{
		{
			Name:      "login",
			Usage:     "登录 LeanCloud 账户",
			Action:    loginAction,
			ArgsUsage: "[account] [password]",
		},
		{
			Name:   "info",
			Usage:  "查看当前登录用户以及应用信息",
			Action: infoAction,
		},
		{
			Name:   "up",
			Usage:  "本地启动云引擎应用",
			Action: upAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "watch",
					Usage: "监听项目文件变更，以自动重启项目",
				},
				cli.IntFlag{
					Name:        "port",
					Usage:       "指定本地调试的端口",
					Value:       3000,
				},
			},
		},
		{
			Name:   "init",
			Usage:  "初始化云引擎项目",
			Action: initAction,
		},
		{
			Name:      "checkout",
			Usage:     "切换当前项目关联的 LeanCloud 应用",
			Action:    checkOutAction,
			ArgsUsage: "[appID]",
		},
		{
			Name:   "deploy",
			Usage:  "部署云引擎项目到服务器",
			Action: deployAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "g",
					Usage:       "从 git 部署项目",
					Destination: &isDeployFromGit,
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "部署当前预备环境的代码至生产环境",
			Action: publishAction,
		},
		{
			Name:      "upload",
			Usage:     "上传文件到当前应用 File 表",
			Action:    uploadAction,
			ArgsUsage: "<file-path>",
		},
		{
			Name:   "env",
			Usage:  "输出运行当前云引擎应用所需要的环境变量",
			Action: envAction,
		},
		{
			Name:      "help",
			Aliases:   []string{"h"},
			Usage:     "显示全部命令或者某个子命令的帮助",
			ArgsUsage: "[command]",
			Action: func(c *cli.Context) error {
				args := c.Args()
				if args.Present() {
					return cli.ShowCommandHelp(c, args.First())
				}

				cli.ShowAppHelp(c)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
