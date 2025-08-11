package main

import (
	"log"

	"github.com/fidrasofyan/s3backup/internal/tasks"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config file",
	Run: func(cmd *cobra.Command, args []string) {
		err := tasks.InitializeConfig()

		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
