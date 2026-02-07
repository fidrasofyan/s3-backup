package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fidrasofyan/db-backup/internal/config"
	"github.com/fidrasofyan/db-backup/internal/service"
)

type backupFile struct {
	Path    string
	ModTime time.Time
	Name    string
}

func DeleteOldBackup(ctx context.Context, cfg *config.Config, storageService *service.Storage, keep int) error {
	if keep <= 0 {
		return nil
	}

	// 1. Scan directory for backup files
	var allFiles []backupFile
	err := filepath.WalkDir(cfg.LocalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Only match *.sql.gz files
		matched, err := filepath.Match("*.sql.gz", d.Name())
		if err != nil {
			return fmt.Errorf("failed to match file: %v", err)
		}
		if !matched {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		allFiles = append(allFiles, backupFile{
			Path:    path,
			ModTime: info.ModTime(),
			Name:    info.Name(),
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan directory: %v", err)
	}

	// 2. Process each database from configuration
	var deletedCounter int32

	for _, dbConfig := range cfg.DBConfigurations {
		var dbFiles []backupFile
		prefix := dbConfig.DBName + "_"

		for _, f := range allFiles {
			// Check if file belongs to this database
			// Format: [dbname]_[timestamp].sql.gz
			if strings.HasPrefix(f.Name, prefix) {
				// Ensure it's exactly [dbname] followed by one underscore and then the timestamp
				// (the timestamp format YYYYMMDD-HHMMSS has no underscores)
				rest := strings.TrimSuffix(f.Name[len(prefix):], ".sql.gz")
				if !strings.Contains(rest, "_") {
					dbFiles = append(dbFiles, f)
				}
			}
		}

		if len(dbFiles) <= keep {
			log.Printf("DB: %s | keep: %d | total files: %d | deleted: 0\n", dbConfig.DBName, keep, len(dbFiles))
			continue
		}

		// Sort files by ModTime descending (newest first)
		sort.Slice(dbFiles, func(i, j int) bool {
			return dbFiles[i].ModTime.After(dbFiles[j].ModTime)
		})

		// Files to delete are from index keep onwards
		filesToDelete := dbFiles[keep:]
		for _, file := range filesToDelete {
			log.Printf("DB: %s | deleting file: %s\n", dbConfig.DBName, file.Path)
			if err := os.Remove(file.Path); err != nil {
				return fmt.Errorf("file %s error: failed to delete from local: %v", file.Path, err)
			}

			// Use relative path to include subdirectories for S3 key
			relPath, err := filepath.Rel(cfg.LocalDir, file.Path)
			if err != nil {
				return fmt.Errorf("file %s error: failed to get relative path: %v", file.Name, err)
			}
			s3Key := fmt.Sprintf("%s/%s", strings.TrimLeft(cfg.RemoteDir, "/"), relPath)

			// Delete file from S3
			err = storageService.Remove(ctx, cfg.AWS.Bucket, s3Key)
			if err != nil {
				log.Printf("Warning: failed to delete %s from S3: %v\n", s3Key, err)
			}

			deletedCounter++
		}
		log.Printf("DB: %s | keep: %d | total files: %d | deleted: %d\n", dbConfig.DBName, keep, len(dbFiles), len(filesToDelete))
	}

	log.Printf("Rotation complete. Total deleted: %d\n", deletedCounter)
	return nil
}
