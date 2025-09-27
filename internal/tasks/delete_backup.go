package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/service"
)

func DeleteOldBackup(ctx context.Context, days int, since time.Time) error {
	storageService, err := service.NewStorage(ctx, &service.NewStorageParams{
		AWSEndpoint:        config.Cfg.AWS.Endpoint,
		AWSRegion:          config.Cfg.AWS.Region,
		AWSAccessKeyID:     config.Cfg.AWS.AccessKeyID,
		AWSSecretAccessKey: config.Cfg.AWS.SecretAccessKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage service: %v", err)
	}

	cutoffTime := since
	if days > 0 {
		cutoffTime = since.AddDate(0, 0, -days)
	}
	var deletedCounter int32

	err = filepath.WalkDir(config.Cfg.LocalDir, func(path string, d os.DirEntry, err error) error {
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

		fileInfo, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}

		// Check if file is older than cutoff time
		if fileInfo.ModTime().Before(cutoffTime) {
			log.Println("deleting file:", path)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("file %s error: failed to delete from local: %v", path, err)
			}

			// Use relative path to include subdirectories
			s3Key, err := filepath.Rel(config.Cfg.LocalDir, path)
			if err != nil {
				return fmt.Errorf("file %s error: failed to get relative path: %v", fileInfo.Name(), err)
			}
			s3Key = fmt.Sprintf("%s/%s", strings.TrimLeft(config.Cfg.RemoteDir, "/"), s3Key)

			// Delete file from S3
			err = storageService.Remove(ctx, config.Cfg.AWS.Bucket, s3Key)
			if err != nil {
				return fmt.Errorf("file %s error: failed to delete from S3: %v", s3Key, err)
			}

			atomic.AddInt32(&deletedCounter, 1)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %v", err)
	}

	log.Printf("keep days: %d | deleted: %d\n", days, deletedCounter)
	return nil
}
