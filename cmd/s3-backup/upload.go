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

var uploadConfigPathFlag string

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload directory contents to S3",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		if uploadConfigPathFlag != "" {
			err := config.LoadConfig(uploadConfigPathFlag)
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

		if err := tasks.Upload(ctx); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	// Flags
	uploadCmd.Flags().StringVarP(&uploadConfigPathFlag, "config", "c", "", "Path to config file. Run 's3-backup init' if you don't have one.")

	rootCmd.AddCommand(uploadCmd)
}
