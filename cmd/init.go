package main

import (
	"github.com/fidrasofyan/s3backup/internal/tasks"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config file",
	Run: func(cmd *cobra.Command, args []string) {
		tasks.InitializeConfig()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
