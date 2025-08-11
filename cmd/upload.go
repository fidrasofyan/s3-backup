package main

import (
	"github.com/fidrasofyan/s3backup/internal/config"
	"github.com/fidrasofyan/s3backup/internal/tasks"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload directory to S3",
	Run: func(cmd *cobra.Command, args []string) {
		tasks.Upload(&tasks.UploadParams{
			AWSEndpoint:        config.Cfg.AWSEndpoint,
			AWSRegion:          config.Cfg.AWSRegion,
			AWSAccessKeyID:     config.Cfg.AWSAccessKeyID,
			AWSAccessSecretKey: config.Cfg.AWSAccessSecretKey,
			AWSBucket:          config.Cfg.AWSBucket,
			LocalDir:           config.Cfg.LocalDir,
			RemoteDir:          config.Cfg.RemoteDir,
		})
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}
