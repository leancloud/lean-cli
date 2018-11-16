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
			Usage:     "Log in to LeanCloud",
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
					Usage: "The LeanCloud region to log in to (e.g., US, CN)",
					Value: "CN",
				},
			},
		},
		{
			Name:      "metric",
			Usage:     "Obtain LeanStorage performance metrics of current project",
			Action:    wrapAction(statusAction),
			ArgsUsage: "[--from fromTime --to toTime --format default|json]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from",
					Usage: "Start date, formatted as YYYY-MM-DD，e.g., 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "End date formated as YYYY-MM-DD，e.g., 1926-08-17",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Output format，'default' or 'json'",
				},
			},
		},
		{
			Name:   "info",
			Usage:  "Show information about the current user and app",
			Action: wrapAction(infoAction),
		},
		{
			Name:   "up",
			Usage:  "Start a development instance locally",
			Action: wrapAction(upAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local port to listen on",
					Value: 3000,
				},
				cli.IntFlag{
					Name:  "console-port,c",
					Usage: "Port of the debug console",
				},
				cli.StringFlag{
					Name:  "cmd",
					Usage: "Command to start the project, other arguments except --console-port are ignored",
				},
			},
		},
		{
			Name:      "init",
			Usage:     "Initialize a LeanEngine project",
			Action:    wrapAction(initAction),
			ArgsUsage: "[dest]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "LeanCloud region for the project",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "LeanEngine group",
				},
			},
		},
		{
			Name:      "switch",
			Usage:     "Change the associated LeanCloud app",
			Action:    wrapAction(switchAction),
			ArgsUsage: "[appID | appName]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: "LeanCloud region",
				},
				cli.StringFlag{
					Name:  "group",
					Usage: "LeanEngine group",
				},
			},
		},
		{
			Name:   "deploy",
			Usage:  "Deploy the project to LeanEngine",
			Action: wrapAction(deployAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "g",
					Usage: "Deploy from git repo",
				},
				cli.BoolFlag{
					Name:  "war",
					Usage: "Deploy .war file for Java project. The first .war file in target/ is used by default",
				},
				cli.BoolFlag{
					Name:  "no-cache",
					Usage: "Force download dependencies",
				},
				cli.StringFlag{
					Name:  "leanignore",
					Usage: "Rule file for ignored files in deployment",
					Value: ".leanignore",
				},
				cli.StringFlag{
					Name:  "message,m",
					Usage: "Comment for this deployment, only applicable when deploying from local files",
				},
				cli.BoolFlag{
					Name: "keep-deploy-file",
				},
				cli.BoolFlag{
					Name: "atomic",
				},
				cli.StringFlag{
					Name: "build-root",
				},
				cli.StringFlag{
					Name:  "revision,r",
					Usage: "Git revision or branch. Only applicable when deploying from Git",
					Value: "master",
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "Publish code from staging to production",
			Action: wrapAction(publishAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "atomic",
				},
			},
		},
		{
			Name:      "upload",
			Usage:     "Upload files to the current project (available in the '_File' class)",
			Action:    uploadAction,
			ArgsUsage: "<file-path> <file-path> ...",
		},
		{
			Name:   "logs",
			Usage:  "Show LeanEngine logs",
			Action: wrapAction(logsAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "Wait for and continuously show most recent logs",
				},
				cli.StringFlag{
					Name:  "env,e",
					Usage: "Environment to view (staging / production)",
					Value: "production",
				},
				cli.IntFlag{
					Name:  "limit,l",
					Usage: "Maximum number of lines to show",
					Value: 30,
				},
				cli.StringFlag{
					Name:  "from",
					Usage: "Start date formatted as YYYY-MM-DD，e.g., 1926-08-17",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "End date formated as YYYY-MM-DD，e.g., 1926-08-17",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Format to use ('default' or 'json')",
					Value: "default",
				},
			},
		},
		{
			Name:   "debug",
			Usage:  "Start the debug console without running the project",
			Action: wrapAction(debugAction),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote,r",
					Usage: "URL of target app",
					Value: "http://localhost:3000",
				},
				cli.StringFlag{
					Name:  "app-id",
					Usage: "Target AppID, use the AppID of the current project by default",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Port to listen on",
					Value: 3001,
				},
			},
		},
		{
			Name:   "env",
			Usage:  "Output environment variables used by the current project",
			Action: wrapAction(envAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local port for the app (affects value of LC_APP_PORT)",
					Value: 3000,
				},
				cli.StringFlag{
					Name:  "template",
					Usage: "Template for output, 'export {{name}}={{value}}' by default",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "Set the value of an environment variable",
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
			Usage:  "LeanCache management commands",
			Action: wrapAction(cacheAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "db",
					Usage: "Name of LeanCache DB",
					Value: -1,
				},
				cli.StringFlag{
					Name:  "name",
					Usage: "Name of LeanCache instance",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "LeanCache command to run",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "Show LeanCache instances of the current project",
					Action: wrapAction(cacheListAction),
				},
			},
		},
		{
			Name:   "cql",
			Usage:  "Start CQL interactive mode",
			Action: wrapAction(cqlAction),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format,f",
					Usage: "CQL result format",
					Value: "table",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "CQL command to run",
				},
			},
		},
		{
			Name:      "search",
			Usage:     "Search development docs",
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
			Usage:     "Show all commands or help info for one command",
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
