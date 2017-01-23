package commands

import (
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/logo"
	"github.com/leancloud/lean-cli/version"
	"github.com/pkg/browser"
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

func Run(args []string) {
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
			ArgsUsage: "[-u username -p password (--region <CN> | <US> | <TAB>)]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username,u",
					Usage: "用户名",
				},
				cli.StringFlag{
					Name:  "password,p",
					Usage: "密码",
				},
				cli.StringFlag{
					Name:  "region,r",
					Usage: "需要登录的节点",
					Value: "CN",
				},
			},
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
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "目标应用节点",
				},
			},
		},
		{
			Name:      "checkout",
			Usage:     "切换当前项目关联的 LeanCloud 应用",
			Action:    checkOutAction,
			ArgsUsage: "[appID | appName]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "目标应用节点",
				},
			},
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
				cli.BoolFlag{
					Name:  "war",
					Usage: "对于 Java 运行环境，直接部署 war 文件。默认部署 target 目录下找到的第一个 war 文件",
				},
				cli.BoolFlag{
					Name:  "no-cache",
					Usage: "强制更新第三方依赖",
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
				cli.BoolFlag{
					Name: "keep-deploy-file",
				},
				cli.StringFlag{
					Name:  "revision,r",
					Usage: "git 的版本号或分支，仅对从 git 仓库部署有效",
					Value: "master",
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
			ArgsUsage: "<file-path> <file-path> ...",
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
				cli.StringFlag{
					Name:  "format",
					Usage: "日志展示格式",
					Value: "default",
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
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "设置新的环境变量",
					Action:    envSetAction,
					ArgsUsage: "[env-name] [env-value]",
				},
				{
					Name:      "unset",
					Usage:     "删除环境变量",
					Action:    envUnsetAction,
					ArgsUsage: "[env-name]",
				},
			},
		},
		{
			Name:   "cache",
			Usage:  "LeanCache 管理相关功能",
			Action: cacheAction,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "db",
					Usage: "需要连接的 LeanCache 实例 db",
					Value: -1,
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "需要连接的 LeanCache 实例名",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "需要立即执行的 LeanCache 命令",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "列出当前应用关联的所有 LeanCache",
					Action: cacheListAction,
				},
			},
		},
		{
			Name:   "cql",
			Usage:  "进入 CQL 交互查询",
			Action: cqlAction,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format,f",
					Usage: "指定 CQL 结果展示格式",
					Value: "table",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "需要立即执行的 CQL 命令",
				},
			},
		},
		{
			Name:      "search",
			Usage:     "根据关键词查询开发文档",
			ArgsUsage: "<kwywords>",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					cli.ShowCommandHelp(c, "search")
				}
				keyword := strings.Join(c.Args(), " ")
				browser.OpenURL("https://leancloud.cn/search.html?q=" + url.QueryEscape(keyword))

				return nil
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

	app.Run(args)
}
