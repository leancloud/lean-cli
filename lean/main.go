package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/getsentry/raven-go"
	"github.com/leancloud/lean-cli/lean/logo"
	"github.com/leancloud/lean-cli/lean/output"
	"github.com/leancloud/lean-cli/lean/stats"
	"github.com/leancloud/lean-cli/lean/version"
)

var (
	op = output.NewOutput(os.Stdout)
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

func run() {
	if len(os.Args) >= 2 && os.Args[1] == "--_collect-stats" {
		stats.Init("Rp8mUcQBVObk8EuyVMDPv39U-gzGzoHsz", "9g3bs563vEsOGdycO2E9ly0y")
		stats.Client.AppVersion = version.Version
		stats.Client.AppChannel = pkgType

		var event string
		if len(os.Args) >= 3 {
			event = os.Args[2]
		}

		stats.Collect([]stats.Event{
			{
				Event: event,
			},
		})
		return
	}

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
					Name:  "port,p",
					Usage: "指定本地调试的端口",
					Value: 3000,
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
					Name:  "g",
					Usage: "从 git 部署项目",
				},
				cli.StringFlag{
					Name:  "leanignore",
					Usage: "部署过程中需要忽略的文件的规则",
					Value: ".leanignore",
				},
				cli.StringFlag{
					Name:  "message,m",
					Usage: "本次部署备注，仅对从本地文件部署项目有效",
					Value: "从命令行工具构建",
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
			Name:   "logs",
			Usage:  "查看 LeanEngine 产生的日志",
			Action: logsAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "持续查看最新日志",
				},
				cli.StringFlag{
					Name:  "env,e",
					Usage: "日志环境，可选项为 staging / production",
					Value: "production",
				},
				cli.IntFlag{
					Name:  "limit,l",
					Usage: "获取日志条目数",
					Value: 30,
				},
			},
		},
		{
			Name:   "env",
			Usage:  "输出运行当前云引擎应用所需要的环境变量",
			Action: envAction,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "指定本地调试的端口",
					Value: 3000,
				},
			},
		},
		{
			Name:   "cache",
			Usage:  "LeanCache 管理相关功能",
			Action: cacheAction,
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "列出当前应用关联的所有 LeanCache",
					Action: cacheListAction,
				},
			},
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

	app.Before = func(c *cli.Context) error {
		args := []string{"--_collect-stats"}
		args = append(args, c.Args()...)
		err := exec.Command(os.Args[0], args...).Start()
		_ = err
		return nil
	}

	app.Run(os.Args)
}

func init() {
	err := raven.SetDSN("https://9cb0f83042044458b2798635c6d9f895:0ff60f888a584fa9918cebc42b09e20d@sentry.avoscloud.com/2")
	if err != nil {
		panic(err)
	}
}

func main() {
	raven.SetTagsContext(map[string]string{
		"version": version.Version,
		"OS":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	})
	err, id := raven.CapturePanicAndWait(run, nil)
	if err != nil {
		fmt.Printf("panic: %s, 错误 ID: %s\r\n", err, id)
		os.Exit(1)
	}
}
