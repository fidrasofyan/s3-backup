package tasks

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fidrasofyan/s3-backup/internal/config"
)

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func BackupDB(ctx context.Context) error {
	// Prerequisites
	var backupCommand string
	if commandExists("mysqldump") {
		backupCommand = "mysqldump"
	} else if commandExists("mariadb-dump") {
		backupCommand = "mariadb-dump"
	} else {
		return fmt.Errorf("mysqldump or mariadb-dump command not found")
	}

	for _, dbConfig := range config.Cfg.BackupDB {
		log.Printf("backing up database: %s:%s/%s\n", dbConfig.Host, dbConfig.Port, dbConfig.DBName)

		// Create file
		filename := fmt.Sprintf(
			"%s/%s_%s.sql.gz",
			config.Cfg.LocalDir,
			dbConfig.DBName,
			time.Now().Format("2006-01-02_15-04-05"),
		)
		file, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("backup db failed: %v", err)
		}
		defer file.Close()

		// Create gzip writer
		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		// Backup command
		cmd := exec.CommandContext(
			ctx,
			backupCommand,
			"--quick",
			"--single-transaction",
			"--routines",
			"--triggers",
			"--events",
			"-h"+dbConfig.Host,
			"-P"+dbConfig.Port,
			"-u"+dbConfig.User,
			"-p"+dbConfig.Password,
			dbConfig.DBName,
		)

		cmd.Stdout = gzipWriter
		cmd.Stderr = os.Stderr

		// Run
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("backup db failed: %v", err)
		}

		// Finalize compression + file
		if err := gzipWriter.Close(); err != nil {
			return fmt.Errorf("closing gzip writer failed: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("closing file failed: %w", err)
		}
	}

	log.Println("backup database complete!")
	return nil
}
