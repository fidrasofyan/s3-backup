package main

import (
	"github.com/fidrasofyan/db-backup/internal/tasks"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		tasks.Version()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
