package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
)

var (
	ErrEmptyFile = errors.New("empty file")
)

type StorageService struct {
	client *s3.Client
}

func (s *StorageService) IsFileExists(ctx context.Context, bucket, key string) (*bool, error) {
	res, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	var nfe *types.NotFound
	if err != nil && !errors.As(err, &nfe) {
		return nil, err
	}

	if res != nil && res.ETag != nil {
		return aws.Bool(true), nil
	}

	return aws.Bool(false), nil
}

func (s *StorageService) Remove(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

type MultipartUploadParams struct {
	PartSize    int64
	Concurrency int
	Bucket      string
	Key         string
	Filepath    string
}

func (s *StorageService) MultipartUpload(ctx context.Context, params *MultipartUploadParams) error {
	// Validate parameters
	if params == nil {
		return errors.New("params cannot be nil")
	}
	if params.Bucket == "" || params.Key == "" || params.Filepath == "" {
		return errors.New("bucket, key, and filepath are required")
	}
	if params.Concurrency <= 0 {
		params.Concurrency = 1
	}
	if params.PartSize <= 0 {
		params.PartSize = 5 * 1024 * 1024 // Default 5MB
	}

	// Open file
	file, err := os.Open(params.Filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	totalSize := fileInfo.Size()
	if totalSize == 0 {
		return ErrEmptyFile
	}

	// For small file, use single part
	if totalSize <= params.PartSize {
		return s.singlePartUpload(ctx, params.Bucket, params.Key, file)
	}

	// Initialize multipart upload
	initResp, err := s.client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(params.Bucket),
		Key:    aws.String(params.Key),
	})
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload: %v", err)
	}

	// Calculate parts
	totalParts := int((totalSize + params.PartSize - 1) / params.PartSize)
	completedParts := make([]types.CompletedPart, totalParts)

	// Upload parts concurrently
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(params.Concurrency)

	for partNum := 1; partNum <= totalParts; partNum++ {
		// Capture loop variables to avoid closure issues
		partNumber := int32(partNum)
		offset := int64(partNumber-1) * params.PartSize
		currentPartSize := params.PartSize
		if offset+params.PartSize > totalSize {
			currentPartSize = totalSize - offset
		}

		g.Go(func() error {
			// Check if context is cancelled
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}

			sectionReader := io.NewSectionReader(file, offset, currentPartSize)

			resp, err := s.client.UploadPart(ctx, &s3.UploadPartInput{
				Bucket:     aws.String(params.Bucket),
				Key:        aws.String(params.Key),
				UploadId:   initResp.UploadId,
				PartNumber: aws.Int32(partNumber),
				Body:       sectionReader,
			})
			if err != nil {
				return fmt.Errorf("failed to upload part %d: %w", partNumber, err)
			}

			completedParts[partNumber-1] = types.CompletedPart{
				ETag:       resp.ETag,
				PartNumber: aws.Int32(partNumber),
			}
			return nil
		})
	}

	// Wait for all uploads or first error
	if err := g.Wait(); err != nil {
		// Abort multipart upload
		s.abortMultipartUpload(params.Bucket, params.Key, initResp.UploadId)
		return fmt.Errorf("multipart upload failed: %v", err)
	}

	// Complete multipart upload
	_, err = s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(params.Bucket),
		Key:      aws.String(params.Key),
		UploadId: initResp.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})

	if err != nil {
		// Try to abort the upload if completion fails
		s.abortMultipartUpload(params.Bucket, params.Key, initResp.UploadId)
		return fmt.Errorf("failed to complete multipart upload: %v", err)
	}

	return nil
}

func (s *StorageService) singlePartUpload(ctx context.Context, bucket, key string, file *os.File) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

func (s *StorageService) abortMultipartUpload(bucket, key string, uploadId *string) {
	abortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := s.client.AbortMultipartUpload(abortCtx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: uploadId,
	})
	if err != nil {
		log.Printf("Warning: failed to abort multipart upload: %v\n", err)
	}
}

type NewStorageServiceParams struct {
	AWSEndpoint        string
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
}

func NewStorageService(ctx context.Context, params *NewStorageServiceParams) (*StorageService, error) {
	// Validate parameters
	if params == nil {
		return nil, errors.New("params cannot be nil")
	}
	if params.AWSEndpoint == "" {
		return nil, errors.New("AWS endpoint is required")
	}
	if params.AWSRegion == "" {
		return nil, errors.New("AWS region is required")
	}
	if params.AWSAccessKeyID == "" || params.AWSSecretAccessKey == "" {
		return nil, errors.New("AWS credentials are required")
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithBaseEndpoint(params.AWSEndpoint),
		config.WithRegion(params.AWSRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				params.AWSAccessKeyID,
				params.AWSSecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &StorageService{
		client: s3Client,
	}, nil
}
