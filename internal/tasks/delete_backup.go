package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/service"
)

type backupFile struct {
	Path    string
	ModTime time.Time
	Name    string
}

func DeleteOldBackup(ctx context.Context, cfg *config.Config, storageService *service.Storage, keepLast int) error {
	if keepLast <= 0 {
		return nil
	}

	var backupFiles []backupFile
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

		backupFiles = append(backupFiles, backupFile{
			Path:    path,
			ModTime: info.ModTime(),
			Name:    info.Name(),
		})
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan directory: %v", err)
	}

	// Sort files by ModTime descending (newest first)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].ModTime.After(backupFiles[j].ModTime)
	})

	if len(backupFiles) <= keepLast {
		log.Printf("keep last: %d | total files: %d | deleted: 0\n", keepLast, len(backupFiles))
		return nil
	}

	// Files to delete are from index keepLast onwards
	filesToDelete := backupFiles[keepLast:]
	var deletedCounter int32

	for _, file := range filesToDelete {
		log.Println("deleting file:", file.Path)
		if err := os.Remove(file.Path); err != nil {
			return fmt.Errorf("file %s error: failed to delete from local: %v", file.Path, err)
		}

		// Use relative path to include subdirectories
		relPath, err := filepath.Rel(cfg.LocalDir, file.Path)
		if err != nil {
			return fmt.Errorf("file %s error: failed to get relative path: %v", file.Name, err)
		}
		s3Key := fmt.Sprintf("%s/%s", strings.TrimLeft(cfg.RemoteDir, "/"), relPath)

		// Delete file from S3
		err = storageService.Remove(ctx, cfg.AWS.Bucket, s3Key)
		if err != nil {
			return fmt.Errorf("file %s error: failed to delete from S3: %v", s3Key, err)
		}

		atomic.AddInt32(&deletedCounter, 1)
	}

	log.Printf("keep last: %d | total files: %d | deleted: %d\n", keepLast, len(backupFiles), deletedCounter)
	return nil
}
