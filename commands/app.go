package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/leancloud/lean-cli/logo"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

// Run the command line
func Run(args []string) {
	// add banner text to help text
	cli.AppHelpTemplate = logo.Logo() + cli.AppHelpTemplate
	cli.SubcommandHelpTemplate = logo.Logo() + cli.SubcommandHelpTemplate

	app := cli.NewApp()
	app.Name = version.Distribution
	app.Version = version.Version
	app.Usage = "Command line to manage and deploy LeanCloud apps"
	app.EnableBashCompletion = true

	app.CommandNotFound = thirdPartyCommand
	app.Commands = []cli.Command{
		{
			Name: "login",
			Usage: func() string {
				if version.Distribution == "lean" {
					return "Log in to LeanCloud"
				} else {
					return "Log in to TapTap Developer Services"
				}
			}(),
			Action: wrapAction(loginAction),
			ArgsUsage: func() string {
				if version.Distribution == "lean" {
					return "[-u <username>] [-p <password>] [--use-token] [--token <token>] [--region (cn-n1 | cn-e1 | us-w1)]"
				} else {
					return "[--token [<token>]]"
				}
			}(),
			Flags: func() []cli.Flag {
				if version.Distribution == "lean" {
					return []cli.Flag{
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
							Usage: "The LeanCloud region to log in to (e.g., cn-n1, us-w1)",
						},
						cli.BoolFlag{
							Name:  "use-token",
							Usage: "Use AccessToken to log in",
						},
						cli.StringFlag{
							Name:  "token",
							Usage: "AccessToken generated from the Dashboard",
						},
					}
				} else {
					return []cli.Flag{
						cli.StringFlag{
							Name:  "region,r",
							Usage: "The TDS region to log in to (e.g., cn-tds1, ap-sg)",
						},
						cli.StringFlag{
							Name:  "token",
							Usage: "AccessToken generated from the Dashboard",
						},
					}
				}
			}(),
		},
		{
			Name: "switch",
			Usage: func() string {
				if version.Distribution == "lean" {
					return "Change the associated LeanCloud app"
				} else {
					return "Change the associated CloudEngine app"
				}
			}(),
			Action:    wrapAction(switchAction),
			ArgsUsage: "[appID | appName]",
			Flags: func() []cli.Flag {
				return []cli.Flag{
					cli.StringFlag{
						Name:  "region",
						Usage: "LeanCloud region",
					},
					cli.StringFlag{
						Name:  "group",
						Usage: "LeanEngine group",
					},
				}
			}(),
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
					Usage: "End date formatted as YYYY-MM-DD，e.g., 1926-08-17",
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
				cli.BoolFlag{
					Name:  "overwrite-functions",
					Usage: "Overwrite cloud functions with the same name in other groups",
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
				cli.StringFlag{
					Name:  "revision,r",
					Usage: "Git revision or branch. Only applicable when deploying from Git",
					Value: "master",
				},
				cli.StringFlag{
					Name:  "options",
					Usage: "Send additional deploy options to server, in urlencode format(like `--options build-root=app&atomic=true`)",
				},
				cli.StringFlag{
					Name:  "prod",
					Usage: "Deploy to production(`--prod 1`) or staging(`--prod 0`) environment, default to staging if it exists",
				},
				cli.BoolFlag{
					Name:  "direct",
					Usage: "Upload project's tarball to remote directly",
				},
			},
		},
		{
			Name:   "publish",
			Usage:  "Publish code from staging to production",
			Action: wrapAction(publishAction),
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "overwrite-functions",
					Usage: "Overwrite cloud functions with the same name in other groups",
				},
				cli.StringFlag{
					Name:  "options",
					Usage: "Send additional deploy options to server, in urlencode format(like `--options build-root=app&atomic=true`)",
				},
			},
		},
		{
			Name:   "db",
			Usage:  "List LeanDB instances under current app (include share instances)",
			Action: wrapAction(dbListAction),
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "List LeanDB instances under current app (include share instances)",
					Action: wrapAction(dbListAction),
				},
				{
					Name:   "proxy",
					Usage:  "Proxy LeanDB instance to local port",
					Action: wrapAction(dbProxyAction),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "app-id",
							Usage: "Specify appId for same name share instance",
						},
						cli.IntFlag{
							Name:  "port, p",
							Usage: "Specify local proxy port",
							Value: 5678,
						},
					},
					ArgsUsage: "<instance-name>",
				},
				{
					Name:   "shell",
					Usage:  "Proxy LeanDB instance to local port and connect using local cli",
					Action: wrapAction(dbShellAction),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "app-id",
							Usage: "Specify appId for same name share instance",
						},
						cli.IntFlag{
							Name:  "port, p",
							Usage: "Specify local proxy port",
							Value: 5678,
						},
					},
					ArgsUsage: "<instance-name>",
				},
			},
		},
		{
			Name:      "upload",
			Usage:     "Upload files to the current application (available in the '_File' class)",
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
					Usage: "Start date formatted as YYYY-MM-DD (local time) or RFC3339，e.g., 2006-01-02 or 2006-01-02T15:04:05+08:00",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "End date formated as YYYY-MM-DD (local time) or RFC3339, e.g., 2006-01-02 or 2006-01-02T15:04:05+08:00",
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
			Usage:  "LeanCache shell",
			Action: wrapAction(cacheAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "db",
					Usage: "Number of LeanCache DB",
					Value: 0,
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
			Usage:  "Start CQL interactive mode (warn: CQL is deprecated)",
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
			Name:    "help",
			Aliases: []string{"h"},
			Usage:   "Show all commands or help info for one command",
			Action: func(c *cli.Context) error {
				args := c.Args()
				if args.Present() {
					_, err := fmt.Printf("Please use `lean %s -h` for subcommand usage.\n", args.First())
					return err
				}
				return cli.ShowAppHelp(c)
			},
		},
	}

	app.Before = func(c *cli.Context) error {
		disableGA, ok := os.LookupEnv("NO_ANALYTICS")
		if !ok || disableGA == "false" {
			args := []string{"--_collect-stats"}
			args = append(args, c.Args()...)
			_ = exec.Command(os.Args[0], args...).Start()
		}
		return nil
	}

	app.Run(args)
}
