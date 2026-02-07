package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/service"
	"github.com/fidrasofyan/s3-backup/internal/tasks"
	"github.com/spf13/cobra"
)

var uploadConfigPathFlag string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload directory contents to S3",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		cfg, err := config.New(uploadConfigPathFlag)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		// Context
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancel()

		// Handle Ctrl+C (SIGINT) and SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			log.Println("Received interrupt signal, exiting...")
			cancel()
		}()

		// Create storage service
		storage, err := service.NewStorage(ctx, &service.NewStorageParams{
			AWSEndpoint:        cfg.AWS.Endpoint,
			AWSRegion:          cfg.AWS.Region,
			AWSAccessKeyID:     cfg.AWS.AccessKeyID,
			AWSSecretAccessKey: cfg.AWS.SecretAccessKey,
		})
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		if err := tasks.Upload(ctx, cfg, storage); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	// Flags
	uploadCmd.Flags().StringVarP(&uploadConfigPathFlag, "config", "c", "", "Path to config file. Run 's3-backup init' if you don't have one.")

	rootCmd.AddCommand(uploadCmd)
}
