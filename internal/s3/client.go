package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const yandexS3Endpoint = "https://storage.yandexcloud.net"

// YandexS3Client wraps the S3 client for Yandex Object Storage
type YandexS3Client struct {
	client *s3.Client
	bucket string
}

// NewYandexS3Client creates a new client for Yandex S3
func NewYandexS3Client(ctx context.Context, bucketName string, env string) (*YandexS3Client, error) {
	keyID := os.Getenv("YANDEX_STORAGE_KEY_ID")
	secretKey := os.Getenv("YANDEX_STORAGE_KEY")

	// If in production environment, panic if credentials are not set
	if env == "prod" && (keyID == "" || secretKey == "") {
		panic("YANDEX_STORAGE_KEY_ID and YANDEX_STORAGE_KEY must be set in production environment")
	}

	// If in development environment and credentials are not set, return nil client
	if env == "dev" && (keyID == "" || secretKey == "") {
		slog.Info("YANDEX_STORAGE_KEY_ID and YANDEX_STORAGE_KEY not set, skipping S3 initialization")
		return nil, nil
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ru-central1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keyID, secretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(yandexS3Endpoint)
		o.UsePathStyle = true
	})

	return &YandexS3Client{
		client: client,
		bucket: bucketName,
	}, nil
}

// DownloadFile downloads a file from Yandex S3
func (s *YandexS3Client) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}
	
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("read object body: %w", err)
	}

	return data, nil
}

// UploadFile uploads data to Yandex S3
func (s *YandexS3Client) UploadFile(
	ctx context.Context,
	key string,
	data []byte,
	contentType string,
	makePublic bool,
) error {
	if s == nil {
		return fmt.Errorf("S3 client not initialized")
	}
	
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if makePublic {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}

	return nil
}

// DeleteFile deletes a file from Yandex S3
func (s *YandexS3Client) DeleteFile(ctx context.Context, key string) error {
	if s == nil {
		return fmt.Errorf("S3 client not initialized")
	}
	
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

// GetPublicFileURL returns a public URL (object must be publicly readable)
func (s *YandexS3Client) GetPublicFileURL(key string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		yandexS3Endpoint,
		s.bucket,
		url.PathEscape(key),
	)
}

// GeneratePresignedUploadURL creates a presigned PUT URL
func (s *YandexS3Client) GeneratePresignedUploadURL(
	ctx context.Context,
	key string,
	contentType string,
	expiration time.Duration,
) (string, error) {
	if s == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}
	
	presigner := s3.NewPresignClient(s.client)

	out, err := presigner.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(key),
			ContentType: aws.String(contentType),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = expiration
		},
	)
	if err != nil {
		return "", fmt.Errorf("presign put object: %w", err)
	}

	return out.URL, nil
}