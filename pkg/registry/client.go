package registry

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
	Opt *Options
	S3  *s3.Client
	Pre *s3.PresignClient
}

// New creates a new registry client
func New(opt *Options) (*Client, error) {
	slog.Debug("Creating registry client")

	if err := opt.Normalize(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if opt.S3Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(opt.S3Endpoint)
			o.UsePathStyle = true
		})
	}

	s3Client := s3.NewFromConfig(awsCfg, s3Opts...)
	preClient := s3.NewPresignClient(s3Client)

	client := &Client{
		Opt: opt,
		S3:  s3Client,
		Pre: preClient,
	}

	// Use background context for connection test during initialization
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	slog.Debug("Registry client ready",
		"endpoint", opt.S3Endpoint,
		"bucket", opt.Bucket,
		"prefix", opt.Prefix)

	return client, nil
}

// Ping verifies S3 connection
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.S3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.Opt.Bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket access: %w", err)
	}
	return nil
}

// Key generates the S3 key path
func (c *Client) Key(sha256 string) string {
	return path.Join(c.Opt.Prefix, sha256)
}

// PutURL generates a presigned URL for upload (default expiry)
func (c *Client) PutURL(ctx context.Context, sha256 string) (string, error) {
	return c.PutURLExpire(ctx, sha256, c.Opt.TTL)
}

// PutURLExpire generates a presigned URL with custom expiry and checksum validation
func (c *Client) PutURLExpire(ctx context.Context, sha256 string, expire time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:            aws.String(c.Opt.Bucket),
		Key:               aws.String(c.Key(sha256)),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256, // Enforce checksum validation
		ChecksumSHA256:    aws.String(sha256),      // Base64-encoded checksum
	}

	res, err := c.Pre.PresignPutObject(ctx, input,
		s3.WithPresignExpires(expire),
	)

	if err != nil {
		return "", fmt.Errorf("presign put: %w", err)
	}

	slog.Debug("Generated PUT URL with checksum validation", 
		"sha256", sha256, 
		"expire", expire)
	return res.URL, nil
}

// GetURL generates a presigned URL for download (default expiry)
func (c *Client) GetURL(ctx context.Context, sha256 string) (string, error) {
	return c.GetURLExpire(ctx, sha256, c.Opt.TTL)
}

// GetURLExpire generates a presigned URL for download with custom expiry
func (c *Client) GetURLExpire(ctx context.Context, sha256 string, expire time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.Opt.Bucket),
		Key:    aws.String(c.Key(sha256)),
	}

	res, err := c.Pre.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expire),
	)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	slog.Debug("Generated GET URL", "sha256", sha256, "expire", expire)
	return res.URL, nil
}
