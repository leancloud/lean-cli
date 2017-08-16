package commands

import (
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/leancloud/lean-cli/logo"
	"github.com/leancloud/lean-cli/version"
	"github.com/pkg/browser"
	"github.com/urfave/cli"
)

// Run the command line
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
			Action:    wrapAction(loginAction),
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
			Name:   "metrix",
			Usage:  "获取当前项目云存储的性能总览",
			Action: wrapAction(statusAction),
			ArgsUsage: "[--from fromTime --to toTime --format default|json]",
			Flags:  []cli.Flag{
				cli.StringFlag{
					Name:  "from",
					Usage: "开始时间，格式为 YYYY-MM-DD，例如 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "结束时间，格式为 YYYY-MM-DD，例如 1926-08-17",
				},
				cli.StringFlag{
					Name: "format",
					Usage: "输出格式，默认为 default，可选 json",
				},
			},
		},
		{
			Name:   "info",
			Usage:  "查看当前登录用户以及应用信息",
			Action: wrapAction(infoAction),
		},
		{
			Name:   "up",
			Usage:  "本地启动云引擎应用",
			Action: wrapAction(upAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "watch",
					Usage: "监听项目文件变更，以自动重启项目",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "指定本地服务端口",
					Value: 3000,
				},
				cli.IntFlag{
					Name:  "console-port,c",
					Usage: "指定调试页面启动端口",
				},
				cli.StringFlag{
					Name:  "cmd",
					Usage: "指定项目启动命令，其他参数将被忽略（--console-port 除外）",
				},
			},
		},
		{
			Name:   "init",
			Usage:  "初始化云引擎项目",
			Action: wrapAction(initAction),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "目标应用节点",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "目标应用 group",
				},
			},
		},
		{
			Name:      "switch",
			Usage:     "切换当前项目关联的 LeanCloud 应用",
			Action:    wrapAction(switchAction),
			ArgsUsage: "[appID | appName]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "目标应用节点",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "目标应用 group",
				},
			},
		},
		{
			Name:      "checkout",
			Usage:     "切换当前项目关联的 LeanCloud 应用",
			Action:    wrapAction(checkOutAction),
			ArgsUsage: "[appID | appName]",
			Hidden:    true,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "目标应用节点",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "目标应用 group",
				},
			},
		},
		{
			Name:   "deploy",
			Usage:  "部署云引擎项目到服务器",
			Action: wrapAction(deployAction),
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
			Action: wrapAction(publishAction),
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
			Action: wrapAction(logsAction),
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
					Name:  "from",
					Usage: "日志开始时间，格式为 YYYY-MM-DD，例如 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "日志结束时间，格式为 YYYY-MM-DD，例如 1926-08-17",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "日志展示格式",
					Value: "default",
				},
			},
		},
		{
			Name:   "debug",
			Usage:  "不运行项目，直接启动云函数调试服务",
			Action: wrapAction(debugAction),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote,r",
					Usage: "目标应用的访问地址，默认为 http://localhost:3000",
					Value: "http://localhost:3000",
				},
				cli.StringFlag{
					Name:  "app-id",
					Usage: "目标应用 appID，如果不指定，则使用当前目录关联应用 appID",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "指定本地调试的端口",
					Value: 3001,
				},
			},
		},
		{
			Name:   "env",
			Usage:  "输出运行当前云引擎应用所需要的环境变量",
			Action: wrapAction(envAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "指定本地调试的端口",
					Value: 3000,
				},
				cli.StringFlag{
					Name:  "template",
					Usage: "指定输出环境变量模版，默认 'export {{name}}={{value}}'",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "设置新的环境变量",
					Action:    wrapAction(envSetAction),
					ArgsUsage: "[env-name] [env-value]",
				},
				{
					Name:      "unset",
					Usage:     "删除环境变量",
					Action:    wrapAction(envUnsetAction),
					ArgsUsage: "[env-name]",
				},
			},
		},
		{
			Name:   "cache",
			Usage:  "LeanCache 管理相关功能",
			Action: wrapAction(cacheAction),
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
					Action: wrapAction(cacheListAction),
				},
			},
		},
		{
			Name:   "cql",
			Usage:  "进入 CQL 交互查询",
			Action: wrapAction(cqlAction),
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
					if err := cli.ShowCommandHelp(c, "search"); err != nil {
						return err
					}
				}
				keyword := strings.Join(c.Args(), " ")
				return browser.OpenURL("https://leancloud.cn/search.html?q=" + url.QueryEscape(keyword))
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
				return cli.ShowAppHelp(c)
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
