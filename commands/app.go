package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

// Run the command line
func Run(args []string) {
	app := cli.NewApp()
	app.Name = version.Distribution
	app.Version = version.Version
	app.Usage = fmt.Sprintf("Command line tool to manage and deploy %s apps", version.EngineBrandName)
	app.EnableBashCompletion = true

	app.CommandNotFound = thirdPartyCommand
	app.Commands = []cli.Command{
		{
			Name:   "login",
			Usage:  fmt.Sprintf("Log in to %s", version.BrandName),
			Action: wrapAction(loginAction),
			ArgsUsage: func() string {
				if version.Distribution == "lean" {
					return "[(-u <username>) (-p <password>) [--use-token] (--token <token>)]"
				} else {
					return "[(--token <token>)]"
				}
			}(),
			Flags: func() []cli.Flag {
				regionsString := []string{}

				for _, r := range version.AvailableRegions {
					regionsString = append(regionsString, r.String())
				}

				regions := strings.Join(regionsString, ", ")

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
							Usage: "The LeanCloud region to log in to (" + regions + ")",
						},
						cli.BoolFlag{
							Name:  "use-token",
							Usage: "Use AccessToken to log in (ask AccessToken)",
						},
						cli.StringFlag{
							Name:  "token",
							Usage: "AccessToken from LeanCloud Console => Account settings => Access tokens: ",
						},
					}
				} else {
					return []cli.Flag{
						cli.StringFlag{
							Name:  "region,r",
							Usage: "The TapTap Developer Services region to log in to (" + regions + ")",
						},
						cli.StringFlag{
							Name:  "token",
							Usage: "AccessToken from TapTap Developer Center => your Game => Game Services => Cloud Services => Cloud Engine => Deploy of your group => Deploy using CLI",
						},
					}
				}
			}(),
		},
		{
			Name:      "switch",
			Usage:     fmt.Sprintf("Change the associated %s app", version.EngineBrandName),
			Action:    wrapAction(switchAction),
			ArgsUsage: "[<appID> | <appName>]",
			Flags: func() []cli.Flag {
				return []cli.Flag{
					cli.StringFlag{
						Name:  "region",
						Usage: fmt.Sprintf("%s region", version.BrandName),
					},
					cli.StringFlag{
						Name:  "group",
						Usage: fmt.Sprintf("%s group", version.EngineBrandName),
					},
				}
			}(),
		},
		{
			Name:   "info",
			Usage:  "Show information about the associated user and app",
			Action: wrapAction(infoAction),
		},
		{
			Name:      "up",
			Usage:     "Start a development instance locally with debug console",
			ArgsUsage: "[(--port <port>) (--console-port <port>) (--cmd <cmd>)]",
			Action:    wrapAction(upAction),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local port to listen on [default: 3000]",
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
				cli.BoolFlag{
					Name:  "fetch-env",
					Usage: "Fetch environment variables from LeanEngine (secret variables not included)",
				},
			},
		},
		{
			Name:      "new",
			Usage:     fmt.Sprintf("Create a new %s project from official examples", version.EngineBrandName),
			Action:    wrapAction(newAction),
			ArgsUsage: "<path>",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "region",
					Usage: fmt.Sprintf("%s region", version.BrandName),
				},
				cli.StringFlag{
					Name:  "group",
					Usage: fmt.Sprintf("%s group", version.EngineBrandName),
				},
			},
		},
		{
			Name:      "deploy",
			Usage:     fmt.Sprintf("Deploy the project to %s", version.EngineBrandName),
			Action:    wrapAction(deployAction),
			ArgsUsage: "(--prod | --staging) [--no-cache --build-logs --overwrite-functions]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "prod",
					Usage: "Deploy to production environment",
				},
				cli.BoolFlag{
					Name:  "staging",
					Usage: "Deploy to staging environment",
				},
				cli.BoolFlag{
					Name:  "build-logs",
					Usage: "Print build logs",
				},
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
					Usage: "Disable building cache",
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
					Usage: "Comment for this version, only applicable when deploying from local files",
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
					Usage: "Send additional deploy options to server, in urlencode format(like `--options build-root=app`)",
				},
				cli.BoolFlag{
					Name:  "direct",
					Usage: "Upload project's tarball to remote directly",
				},
			},
		},
		{
			Name:      "preview",
			Usage:     "Manage preview environments",
			ArgsUsage: "(deploy | delete)",
			Subcommands: []cli.Command{
				{
					Name:   "deploy",
					Usage:  "Deploy to preview environment",
					Action: wrapAction(deployPreviewAction),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Name of the preview environment. Will be created if it does not exist.",
						},
						cli.StringFlag{
							Name:  "url",
							Usage: "Pull/Merge request URL",
						},
						cli.StringFlag{
							Name:  "commit",
							Usage: "Commit hash",
						},
						cli.BoolFlag{
							Name:  "build-logs",
							Usage: "Print build logs",
						},
						cli.BoolFlag{
							Name:  "war",
							Usage: "Deploy .war file for Java project. The first .war file in target/ is used by default",
						},
						cli.StringFlag{
							Name:  "leanignore",
							Usage: "Rule file for ignored files in deployment",
							Value: ".leanignore",
						},
					},
				},
				{
					Name:   "delete",
					Usage:  "Delete preview environment",
					Action: wrapAction(deletePreviewAction),
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Usage: "Name of the preview environment",
						},
					},
				},
			},
		},
		{
			Name:      "publish",
			Usage:     "Publish the version of staging to production",
			Action:    wrapAction(publishAction),
			ArgsUsage: "[--overwrite-functions]",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "overwrite-functions",
					Usage: "Overwrite cloud functions with the same name in other groups",
				},
				cli.StringFlag{
					Name:  "options",
					Usage: "Send additional deploy options to server, in urlencode format(like `--options build-root=app`)",
				},
			},
		},
		{
			Name:      "db",
			Usage:     fmt.Sprintf("Access to to %s instances", version.DBBrandName),
			Action:    wrapAction(dbListAction),
			ArgsUsage: "(list | proxy | shell | exec)",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  fmt.Sprintf("List %s instances of current app (include share instances)", version.DBBrandName),
					Action: wrapAction(dbListAction),
				},
				{
					Name:      "proxy",
					Usage:     fmt.Sprintf("Proxy %s instance to local port", version.DBBrandName),
					Action:    wrapAction(dbProxyAction),
					ArgsUsage: "<instance-name>",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "app-id",
							Usage: "Specify appId for same name share instance",
						},
						cli.IntFlag{
							Name:  "port, p",
							Usage: "Specify local proxy port [default: 5678]",
							Value: 5678,
						},
					},
				},
				{
					Name:      "shell",
					Usage:     fmt.Sprintf("Enter interactive shell to %s instance", version.DBBrandName),
					Action:    wrapAction(dbShellAction),
					ArgsUsage: "<instance-name>",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "app-id",
							Usage: "Specify appId for same name share instance",
						},
						cli.IntFlag{
							Name:  "port, p",
							Usage: "Specify local proxy port [default: 5678]",
							Value: 5678,
						},
					},
				},
				{
					Name:      "exec",
					Usage:     fmt.Sprintf("Exce commands on %s instance", version.DBBrandName),
					Action:    wrapAction(dbExecAction),
					ArgsUsage: "<instance-name> <db-commands>...",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "app-id",
							Usage: "Specify appId for same name share instance",
						},
						cli.IntFlag{
							Name:  "port, p",
							Usage: "Specify local proxy port [default: 5678]",
							Value: 5678,
						},
					},
				},
			},
		},
		{
			Name:      "file",
			Usage:     "Manage files ('_File' class in Data Storage)",
			ArgsUsage: "(upload)",
			Subcommands: []cli.Command{
				{
					Name:      "upload",
					Usage:     "Upload files",
					Action:    wrapAction(uploadAction),
					ArgsUsage: "<file-path>...",
				},
			},
		},
		{
			Name:      "logs",
			Usage:     fmt.Sprintf("Show %s logs", version.EngineBrandName),
			Action:    wrapAction(logsAction),
			ArgsUsage: "[-f (--env staging|production)]",
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
					Usage: "Maximum number of lines to show [default: 30]",
					Value: 30,
				},
				cli.StringFlag{
					Name:  "from",
					Usage: "Start date formatted as YYYY-MM-DD (local time) or RFC3339, e.g., 2006-01-02 or 2006-01-02T15:04:05+08:00",
				},
				cli.StringFlag{
					Name:  "to",
					Usage: "End date formatted as YYYY-MM-DD (local time) or RFC3339, e.g., 2006-01-02 or 2006-01-02T15:04:05+08:00",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "Format to use ('default' or 'json')",
					Value: "default",
				},
			},
		},
		{
			Name:      "debug",
			Usage:     "Start the debug console without running the project",
			Action:    wrapAction(debugAction),
			ArgsUsage: "[(--remote <url>) (--port <port>)]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "remote,r",
					Usage: "URL of target app running on",
					Value: "http://localhost:3000",
				},
				cli.StringFlag{
					Name:  "app-id",
					Usage: "Target App ID, use the App ID of the current project by default",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Port to listen on [default: 3001]",
					Value: 3001,
				},
			},
		},
		{
			Name:      "env",
			Usage:     fmt.Sprintf("Print custom environment variables on %s (secret variables not included)", version.EngineBrandName),
			Action:    wrapAction(envAction),
			ArgsUsage: "[(set | unset)]",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port,p",
					Usage: "Local port for the app (affects value of LEANCLOUD_APP_PORT) [default: 3000]",
					Value: 3000,
				},
				cli.StringFlag{
					Name:  "template",
					Usage: "Template for output [default: export {{{name}}}={{{value}}}]",
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "set",
					Usage:     "Set the value of an environment variable",
					Action:    wrapAction(envSetAction),
					ArgsUsage: "<name> <value>",
				},
				{
					Name:      "unset",
					Usage:     "Delete an environment variable",
					Action:    wrapAction(envUnsetAction),
					ArgsUsage: "<name>",
				},
			},
		},
		{
			Name:      "cql",
			Usage:     "Enter CQL interactive shell (warn: CQL is deprecated)",
			Action:    wrapAction(cqlAction),
			ArgsUsage: "[(--format json) (--eval <cql>)]",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "format,f",
					Usage: "CQL result format can be 'table' or 'json' [default: table]",
					Value: "table",
				},
				cli.StringFlag{
					Name:  "eval",
					Usage: "CQL command to run",
				},
			},
		},
		{
			Name:  "help",
			Usage: "Show usages of all subcommands",
			Action: func(c *cli.Context) error {
				args := c.Args()
				if args.Present() {
					_, err := fmt.Printf("Please use `%s %s --help` for subcommand usage.\n", os.Args[0], args.First())
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
			StartBackgroundCommand(exec.Command(os.Args[0], args...))
		}
		return nil
	}

	app.Run(args)
}
