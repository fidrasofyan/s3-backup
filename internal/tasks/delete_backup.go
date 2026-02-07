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

	// 2. Group files by database name
	// The filename format is: [DBName]_[YYYYMMDD]-[HHMMSS].sql.gz
	// Extract [DBName] by finding the last underscore and removing that suffix.
	filesByDB := make(map[string][]backupFile)
	for _, f := range allFiles {
		dbName := extractDBName(f.Name)
		filesByDB[dbName] = append(filesByDB[dbName], f)
	}

	// 3. For each database, keep only the last N backups
	var deletedCounter int32
	var totalFiles int32

	for dbName, files := range filesByDB {
		totalFiles += int32(len(files))

		if len(files) <= keep {
			log.Printf("DB: %s | keep: %d | total files: %d | deleted: 0\n", dbName, keep, len(files))
			continue
		}

		// Sort files by ModTime descending (newest first)
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime.After(files[j].ModTime)
		})

		// Files to delete are from index keepLast onwards
		filesToDelete := files[keep:]
		for _, file := range filesToDelete {
			log.Printf("DB: %s | deleting file: %s\n", dbName, file.Path)
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
				// We don't want to stop everything if S3 delete fails (maybe it was already deleted or never uploaded)
				log.Printf("Warning: failed to delete %s from S3: %v\n", s3Key, err)
			}

			deletedCounter++
		}
		log.Printf("DB: %s | keep: %d | total files: %d | deleted: %d\n", dbName, keep, len(files), len(filesToDelete))
	}

	log.Printf("Total files scanned: %d | Total deleted: %d\n", totalFiles, deletedCounter)
	return nil
}

// extractDBName extracts the database name from the backup filename.
// Format: [DBName]_YYYYMMDD-HHMMSS.sql.gz
func extractDBName(filename string) string {
	// Remove .sql.gz suffix
	base := strings.TrimSuffix(filename, ".sql.gz")

	// Split by underscore. The last part is YYYYMMDD-HHMMSS
	parts := strings.Split(base, "_")
	if len(parts) < 2 {
		return base
	}

	// Join everything except the last part (the timestamp)
	return strings.Join(parts[:len(parts)-1], "_")
}
