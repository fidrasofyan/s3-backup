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

var (
	backupDBConfigPathFlag string
	backupDBNoUploadFlag   bool
	backupDBDeleteDaysFlag int
)

var backupDBCmd = &cobra.Command{
	Use:   "backup-db",
	Short: "Backup database and upload to S3",
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		if backupDBConfigPathFlag != "" {
			config.MustLoadConfig(backupDBConfigPathFlag)
		} else {
			config.MustLoadConfig("")
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
		err := tasks.BackupDB(ctx)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		// Delete old backup
		if backupDBDeleteDaysFlag > 0 {
			err := tasks.DeleteOldBackup(ctx, backupDBDeleteDaysFlag)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}

		// Upload
		if !backupDBNoUploadFlag {
			err := tasks.Upload(ctx)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	},
}

func init() {
	// Flags
	backupDBCmd.Flags().StringVarP(&backupDBConfigPathFlag, "config", "c", "", "Path to config file. Run 's3-backup init' if you don't have one.")
	backupDBCmd.Flags().BoolVar(&backupDBNoUploadFlag, "no-upload", false, "Don't upload to S3")
	backupDBCmd.Flags().IntVar(&backupDBDeleteDaysFlag, "delete-days", 0, "Delete old backups older than N days (file ends with *.sql.gz)")

	rootCmd.AddCommand(backupDBCmd)
}
