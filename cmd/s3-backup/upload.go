package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/tasks"
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

		// Context
		rootCtx, rootCancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer rootCancel()

		// Handle Ctrl+C (SIGINT) and SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			log.Println("Received interrupt signal, exiting...")
			rootCancel()
		}()

		err := tasks.Upload(rootCtx, &tasks.UploadParams{
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
	uploadCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file. Run 's3-backup init' if you don't have one.")

	rootCmd.AddCommand(uploadCmd)
}
