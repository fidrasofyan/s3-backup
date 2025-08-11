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

func Upload(params *UploadParams) {
	ctx := context.Background()

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
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Configure multipart uploader
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024
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
		log.Fatalf("Failed to scan directory: %v", err)
	}

	// Upload files concurrently
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, fileInfo := range files {
		wg.Add(1)

		go func(fi *FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			s3Key, err := filepath.Rel(params.LocalDir, fi.Path)
			if err != nil {
				log.Fatalf("Failed to get relative path: %v", err)
			}
			s3Key = fmt.Sprintf("%s/%s", strings.TrimLeft(params.RemoteDir, "/"), s3Key)

			multipartUpload(ctx, s3Client, uploader, &params.AWSBucket, fi, &s3Key)
		}(&fileInfo)
	}

	wg.Wait()
	log.Println("✅ Upload completed")
}

func isFileExistsInS3(ctx context.Context, s3Client *s3.Client, bucket *string, key *string) bool {
	_, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: bucket,
		Key:    key,
	})
	return err == nil
}

func multipartUpload(
	ctx context.Context,
	s3Client *s3.Client,
	uploader *manager.Uploader,
	bucket *string,
	fileInfo *FileInfo,
	s3Key *string,
) {
	// Is file exists in S3?
	if isFileExistsInS3(ctx, s3Client, bucket, s3Key) {
		log.Printf("Already exists in S3 (skipping): %s", *s3Key)
		return
	}

	file, err := os.Open(fileInfo.Path)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Upload file
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    s3Key,
		Body:   file,
	})

	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	log.Printf("Uploaded: %s ", *s3Key)
}
