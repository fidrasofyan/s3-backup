package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fidrasofyan/s3-backup/internal/config"
	"github.com/fidrasofyan/s3-backup/internal/service"
	"golang.org/x/sync/errgroup"
)

type FileInfo struct {
	Name string
	Size int64
	Path string
}

func Upload(ctx context.Context) error {
	// Create storage service
	storageService, err := service.NewStorageService(ctx, &service.NewStorageServiceParams{
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

	err = filepath.Walk(config.Cfg.LocalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, FileInfo{
			Name: info.Name(),
			Size: info.Size(),
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

	for _, fileInfo := range files {
		fi := fileInfo

		g.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

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
				return nil
			}

			// Upload file
			err = storageService.MultipartUpload(ctx, &service.MultipartUploadParams{
				PartSize:    5 * 1024 * 1024, // 5 MB
				Concurrency: 5,
				Bucket:      config.Cfg.AWS.Bucket,
				Key:         s3Key,
				Filepath:    fi.Path,
			})

			if err != nil {
				if errors.Is(err, service.ErrEmptyFile) {
					log.Printf("file is empty (skipped): %s\n", fileInfo.Path)
					return nil
				}
				return fmt.Errorf("failed to upload file %v: %v", fileInfo.Path, err)
			}

			log.Printf("file uploaded: %s", s3Key)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	log.Println("upload complete!")
	return nil
}
