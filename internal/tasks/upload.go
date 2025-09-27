package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/service"
	"golang.org/x/sync/errgroup"
)

type FileInfo struct {
	Name string
	Path string
}

func Upload(ctx context.Context) error {
	// Create storage service
	storageService, err := service.NewStorage(ctx, &service.NewStorageParams{
		AWSEndpoint:        config.Cfg.AWS.Endpoint,
		AWSRegion:          config.Cfg.AWS.Region,
		AWSAccessKeyID:     config.Cfg.AWS.AccessKeyID,
		AWSSecretAccessKey: config.Cfg.AWS.SecretAccessKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create storage service: %v", err)
	}

	// Scan directory
	files := []FileInfo{}

	err = filepath.WalkDir(config.Cfg.LocalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, FileInfo{
			Name: d.Name(),
			Path: path,
		})
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %v", err)
	}

	// Upload files concurrently
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	var (
		uploadedCounter int32
		skippedCounter  int32
	)

	for _, fileInfo := range files {
		fi := fileInfo

		g.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			log.Printf("uploading file: %s\n", fi.Path)

			// Use relative path to include subdirectories
			s3Key, err := filepath.Rel(config.Cfg.LocalDir, fi.Path)
			if err != nil {
				return fmt.Errorf("file %s error: failed to get relative path: %v", fi.Name, err)
			}
			s3Key = fmt.Sprintf("%s/%s", strings.TrimLeft(config.Cfg.RemoteDir, "/"), s3Key)

			// Is file exists in S3?
			exists, err := storageService.IsFileExists(ctx, config.Cfg.AWS.Bucket, s3Key)
			if err != nil {
				return fmt.Errorf("failed to check if file exists: %v", err)
			}
			if *exists {
				atomic.AddInt32(&skippedCounter, 1)
				return nil
			}

			// Upload file
			err = storageService.Upload(ctx, &service.UploadParams{
				PartSize:    5 * 1024 * 1024, // 5 MB
				Concurrency: 5,
				Bucket:      config.Cfg.AWS.Bucket,
				Key:         s3Key,
				Filepath:    fi.Path,
			})

			if err != nil {
				if errors.Is(err, service.ErrEmptyFile) {
					log.Printf("file is empty (skipped): %s\n", fi.Path)

					atomic.AddInt32(&skippedCounter, 1)
					return nil
				}
				return fmt.Errorf("failed to upload file %v: %v", fi.Path, err)
			}

			log.Printf("file uploaded: %s", s3Key)

			atomic.AddInt32(&uploadedCounter, 1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	log.Printf("uploaded: %d | skipped: %d\n", uploadedCounter, skippedCounter)
	return nil
}
