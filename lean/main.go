package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "lean",
		Short: "LeanEngine command line",
	}

	var runtime string
	var createAppCmd = &cobra.Command{
		Use:   "new [appname] --runtime=[runtime]",
		Short: "Create a new LeanEngine app",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				println("Error: Please specific your app name")
				println("leanengine new [appname] --runtime=[runtime]")
				os.Exit(1)
			}

			appName := args[0]

			if runtime == "python" {
			} else if runtime == "node" || runtime == "nodejs" {
			} else {
				fmt.Println("Error: Please specific a project runtime ( python / node)")
				fmt.Println("leanengine new", appName, "--runtime=[runtime]")
				os.Exit(1)
			}

			if err := os.Mkdir(appName, 0777); os.IsExist(err) {
				println("'" + appName + "' is already exist.")
				os.Exit(1)
			} else if err != nil {
				fmt.Println("Error: ", err)
			}
		},
	}
	createAppCmd.Flags().StringVarP(&runtime, "runtime", "r", "", "project runtime")
	rootCmd.AddCommand(createAppCmd)

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init LeanEngine app in current directory",
		Run: func(cmd *cobra.Command, args []string) {
			err := newAppInfoFromInput()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
		},
	}
	rootCmd.AddCommand(initCmd)

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run LeanEngine app in local",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("run...")
		},
	}
	rootCmd.AddCommand(runCmd)

	var infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Show current app info",
		Run: func(cmd *cobra.Command, args []string) {
			err := printAppInfoFromLocal()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(4)
			}
		},
	}
	rootCmd.AddCommand(infoCmd)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("1.0.0")
		},
	}
	rootCmd.AddCommand(versionCmd)

	rootCmd.Execute()
}
