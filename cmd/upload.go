package main

import (
	"log"

	"github.com/fidrasofyan/s3backup/internal/config"
	"github.com/fidrasofyan/s3backup/internal/tasks"
	"github.com/spf13/cobra"
)

var configPath string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload directory contents to S3",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		if configPath != "" {
			err := config.LoadConfig(configPath)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			err := config.LoadConfig("")
			if err != nil {
				log.Fatalln(err)
			}
		}

		err := tasks.Upload(&tasks.UploadParams{
			AWSEndpoint:        config.Cfg.AWSEndpoint,
			AWSRegion:          config.Cfg.AWSRegion,
			AWSAccessKeyID:     config.Cfg.AWSAccessKeyID,
			AWSAccessSecretKey: config.Cfg.AWSAccessSecretKey,
			AWSBucket:          config.Cfg.AWSBucket,
			LocalDir:           config.Cfg.LocalDir,
			RemoteDir:          config.Cfg.RemoteDir,
		})

		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	// Flags
	uploadCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file. Run 's3backup init' if you don't have one.")

	rootCmd.AddCommand(uploadCmd)
}
