package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "lean",
		Short: "LeanEngine command line tool",
	}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(VERSION)
		},
	}
	rootCmd.AddCommand(versionCmd)

	rootCmd.Execute()
}
