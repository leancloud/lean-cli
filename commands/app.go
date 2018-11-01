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
			Usage:     "Login to LeanCloud",
			Action:    wrapAction(loginAction),
			ArgsUsage: "[-u username -p password (--region <CN> | <US> | <TAB>)]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "username,u",
					Usage: "Username",
				},
				cli.StringFlag{
					Name:  "password,p",
					Usage: "Password",
				},
				cli.StringFlag{
					Name:  "region,r",
					Usage: "Region",
					Value: "CN",
				},
			},
		},
		{
			Name:      "metric",
			Usage:     "Storage service performance metrics of current project",
			Action:    wrapAction(statusAction),
			ArgsUsage: "[--from fromTime --to toTime --format default|json]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from",
					Usage: "Starting time formatted as YYYY-MM-DD, e.g. 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "Ending time formatted as YYYY-MM-DD, e.g. 1926-08-17",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Output format, default or json",
				},
			},
		},
		{
			Name:   "info",
			Usage:  "Current user and app information",
			Action: wrapAction(infoAction),
		},
		{
			Name:   "up",
			Usage:  "Start a development instance locally",
			Action: wrapAction(upAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "watch",
					Usage: "Watch project files and auto-restart on change",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local port to listen",
					Value: 3000,
				},
				cli.IntFlag{
					Name:  "console-port,c",
					Usage: "Debug console port",
				},
				cli.StringFlag{
					Name:  "cmd",
					Usage: "Custom command to run, other arguments will be ignored (except --console-port)",
				},
			},
		},
		{
			Name:      "init",
			Usage:     "Initialize LeanEngine project",
			Action:    wrapAction(initAction),
			ArgsUsage: "[dest]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "Target region",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "Target group",
				},
			},
		},
		{
			Name:      "switch",
			Usage:     "Change the LeanCloud app associated with current project",
			Action:    wrapAction(switchAction),
			ArgsUsage: "[appID | appName]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "Target region",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "Target group",
				},
			},
		},
		{
			Name:      "checkout",
			Usage:     "Change the LeanCloud app associated with current project",
			Action:    wrapAction(checkOutAction),
			ArgsUsage: "[appID | appName]",
			Hidden:    true,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "Target region",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "Target group",
				},
			},
		},
		{
			Name:   "deploy",
			Usage:  "Deploy project to the cloud",
			Action: wrapAction(deployAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "g",
					Usage: "Deploy from git",
				},
				cli.BoolFlag{
					Name:  "war",
					Usage: "Deploy the first .war file in the target directory for a Java project",
				},
				cli.BoolFlag{
					Name:  "no-cache",
					Usage: "Force-update 3rd-party dependencies",
				},
				cli.StringFlag{
					Name:  "leanignore",
					Usage: "Rule file to ignore files during deployment",
					Value: ".leanignore",
				},
				cli.StringFlag{
					Name:  "message,m",
					Usage: "Comments for this deployment. Only applicable when deploying from local files.",
				},
				cli.BoolFlag{
					Name: "keep-deploy-file",
				},
				cli.BoolFlag{
					Name: "atomic",
				},
				cli.StringFlag{
					Name:  "revision,r",
					Usage: "git revision or branch. Only applicable when deploying from git repo",
					Value: "master",
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "Deploy from staging instance to production instances",
			Action: wrapAction(publishAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "atomic",
				},
			},
		},
		{
			Name:      "upload",
			Usage:     "Upload file to the File class of the current app",
			Action:    uploadAction,
			ArgsUsage: "<file-path> <file-path> ...",
		},
		{
			Name:   "logs",
			Usage:  "View LeanEngine logs",
			Action: wrapAction(logsAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "Continuously show updated logs",
				},
				cli.StringFlag{
					Name:  "env,e",
					Usage: "Choose from staging / production",
					Value: "production",
				},
				cli.IntFlag{
					Name:  "limit,l",
					Usage: "Maximum log lines to show",
					Value: 30,
				},
				cli.StringFlag{
					Name:  "from",
					Usage: "Starting time formated as YYYY-MM-DD, e.g. 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "Ending time formated as YYYY-MM-DD, e.g. 1926-08-17",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Log display format",
					Value: "default",
				},
			},
		},
		{
			Name:   "debug",
			Usage:  "Only start cloud function debug console",
			Action: wrapAction(debugAction),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote,r",
					Usage: "URL of target app, default to http://localhost:3000",
					Value: "http://localhost:3000",
				},
				cli.StringFlag{
					Name:  "app-id",
					Usage: "appID of target app, default to the current associated appID",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Port for local debugging",
					Value: 3001,
				},
			},
		},
		{
			Name:   "env",
			Usage:  "Print the environment variables need by the current project",
			Action: wrapAction(envAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local debugging port",
					Value: 3000,
				},
				cli.StringFlag{
					Name:  "template",
					Usage: "Template for printing environment variables. Default: 'export {{name}}={{value}}'",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "Set an environment variable",
					Action:    wrapAction(envSetAction),
					ArgsUsage: "[env-name] [env-value]",
				},
				{
					Name:      "unset",
					Usage:     "Delete an environment variable",
					Action:    wrapAction(envUnsetAction),
					ArgsUsage: "[env-name]",
				},
			},
		},
		{
			Name:   "cache",
			Usage:  "LeanCache management functions",
			Action: wrapAction(cacheAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "db",
					Usage: "LeanCache DB to connect to",
					Value: -1,
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "LeanCache instance to connect to",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "LeanCache command to execute",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List all LeanCache instances associated with current project",
					Action: wrapAction(cacheListAction),
				},
			},
		},
		{
			Name:   "cql",
			Usage:  "Enter interactive CQL query",
			Action: wrapAction(cqlAction),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format,f",
					Usage: "Format of CQL results",
					Value: "table",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "CQL command to execute",
				},
			},
		},
		{
			Name:      "search",
			Usage:     "Search the developer docs for keywords",
			ArgsUsage: "<keywords>",
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
			Usage:     "Display all help information or for given command",
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
