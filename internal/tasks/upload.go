package tasks

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadParams struct {
	AWSEndpoint        string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSAccessSecretKey string
	AWSBucket          string
	LocalDir           string
	RemoteDir          string
}

type FileInfo struct {
	Name string
	Size int64
	Path string
}

func Upload(ctx context.Context, params *UploadParams) error {
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithBaseEndpoint(params.AWSEndpoint),
		config.WithRegion(params.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				params.AWSAccessKeyID,
				params.AWSAccessSecretKey,
				"",
			),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Configure multipart uploader
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10 MB
		u.Concurrency = 5
		u.LeavePartsOnError = false
	})

	// Scan directory
	files := []FileInfo{}

	err = filepath.Walk(params.LocalDir, func(path string, info os.FileInfo, err error) error {
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
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // limit concurrency

	for _, fileInfo := range files {
		// Check if context is done
		select {
		case <-ctx.Done():
			log.Println("operation canceled or timed out")
			return ctx.Err()
		default:
		}

		wg.Add(1)

		go func(fi FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s3Key, err := filepath.Rel(params.LocalDir, fi.Path)
			if err != nil {
				fmt.Printf("file %s error: failed to get relative path: %v \n", fi.Name, err)
				return
			}
			s3Key = fmt.Sprintf("%s/%s", strings.TrimLeft(params.RemoteDir, "/"), s3Key)

			log.Printf("uploading file: %s\n", s3Key)
			err = multipartUpload(ctx, s3Client, uploader, &params.AWSBucket, &fi, &s3Key)
			if err != nil {
				log.Println(err)
			}
		}(fileInfo)
	}

	wg.Wait()
	log.Println("upload completed!")

	return nil
}

func multipartUpload(
	ctx context.Context,
	s3Client *s3.Client,
	uploader *manager.Uploader,
	bucket *string,
	fileInfo *FileInfo,
	s3Key *string,
) error {
	// Is file exists in S3?
	res, _ := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: bucket,
		Key:    s3Key,
	})
	if res != nil && res.ETag != nil {
		return fmt.Errorf("file already exists (skipping): %s", *s3Key)
	}

	file, err := os.Open(fileInfo.Path)
	if err != nil {
		return fmt.Errorf("failed to open file %v: %v", fileInfo.Path, err)
	}
	defer file.Close()

	// Upload file
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    s3Key,
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file %v: %v", fileInfo.Path, err)
	}

	log.Printf("file uploaded: %s", *s3Key)

	return nil
}
