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

func BackupDB(ctx context.Context, cfg *config.Config) error {
	// Prerequisites
	var backupCommand string
	if commandExists("mysqldump") {
		backupCommand = "mysqldump"
	} else if commandExists("mariadb-dump") {
		backupCommand = "mariadb-dump"
	} else {
		return fmt.Errorf("mysqldump or mariadb-dump command not found")
	}

	for _, dbConfig := range cfg.BackupDB {
		if err := backupSingleDB(ctx, backupCommand, cfg, dbConfig); err != nil {
			return err
		}
	}

	log.Println("backup database complete!")
	return nil
}

func backupSingleDB(ctx context.Context, backupCommand string, cfg *config.Config, dbConfig config.BackupDBConfig) error {
	log.Printf("backing up database: %s:%s/%s\n", dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	// Create file
	filename := fmt.Sprintf(
		"%s/%s_%s.sql.gz",
		cfg.LocalDir,
		dbConfig.DBName,
		time.Now().Format("20060102_150405"),
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
		dbConfig.DBName,
	)

	// Pass password via environment variable (more secure)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+dbConfig.Password)
	cmd.Stdout = gzipWriter
	cmd.Stderr = os.Stderr

	// Run
	if err := cmd.Run(); err != nil {
		// Cleanup: close writers and remove partial file
		gzipWriter.Close()
		file.Close()
		os.Remove(filename)
		return fmt.Errorf("backup db failed: %v", err)
	}

	return nil
}
