package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BucketUploader struct {
	client *s3.Client
	bucket string
}

// NewBucketUploader configures an S3-compatible client (Railway Buckets / Tigris).
func NewBucketUploader(endpoint, region, bucket, accessKeyID, secretAccessKey string) *BucketUploader {
	client := s3.New(s3.Options{
		Region:       region,
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	})

	return &BucketUploader{client: client, bucket: bucket}
}

// Upload writes data to <bucket>/<key>, overwriting if it already exists.
func (b *BucketUploader) Upload(ctx context.Context, key string, data []byte) error {
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(b.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("text/csv"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %q to bucket %q: %w", key, b.bucket, err)
	}
	return nil
}
