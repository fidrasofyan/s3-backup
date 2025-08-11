package main

import (
	"log"

	"github.com/fidrasofyan/s3backup/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "s3backup",
	Short: "Backup directory to S3",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Use == "init" {
			return
		}
		config.LoadConfig()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
