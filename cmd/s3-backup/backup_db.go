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

var (
	backupDBConfigPathFlag string
	backupDBNoUploadFlag   bool
	backupDBDeleteDaysFlag int
)

var backupDBCmd = &cobra.Command{
	Use:   "backup-db",
	Short: "Backup database and upload to S3",
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		// Load config
		cfg, err := config.New(backupDBConfigPathFlag)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		// Context
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancel()

		// Setup signal catching
		quitCh := make(chan os.Signal, 1)
		signal.Notify(quitCh,
			os.Interrupt,    // SIGINT (Ctrl+C)
			syscall.SIGTERM, // stop
			syscall.SIGQUIT, // Ctrl+\
			syscall.SIGHUP,  // terminal hangup
		)
		go func() {
			log.Printf("Signal caught: %s", <-quitCh)
			cancel()
		}()

		// Start backup
		if err := tasks.BackupDB(ctx, cfg); err != nil {
			log.Fatalf("Error: %v", err)
		}

		// Create storageService service for S3 operations
		storageService, err := service.NewStorage(ctx, &service.NewStorageParams{
			AWSEndpoint:        cfg.AWS.Endpoint,
			AWSRegion:          cfg.AWS.Region,
			AWSAccessKeyID:     cfg.AWS.AccessKeyID,
			AWSSecretAccessKey: cfg.AWS.SecretAccessKey,
		})
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		// Delete old backup
		if backupDBDeleteDaysFlag >= 0 {
			if err := tasks.DeleteOldBackup(ctx, cfg, storageService, backupDBDeleteDaysFlag, startTime); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}

		// Upload
		if !backupDBNoUploadFlag {
			if err := tasks.Upload(ctx, cfg, storageService); err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	},
}

func init() {
	// Flags
	backupDBCmd.Flags().StringVarP(&backupDBConfigPathFlag, "config", "c", "", "Path to config file. Run 's3-backup init' if you don't have one.")
	backupDBCmd.Flags().BoolVar(&backupDBNoUploadFlag, "no-upload", false, "Don't upload to S3")
	backupDBCmd.Flags().IntVar(&backupDBDeleteDaysFlag, "delete-days", -1, "Delete old backups older than N days (file ends with *.sql.gz)")

	rootCmd.AddCommand(backupDBCmd)
}
