package registry

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	CFG Config
	S3  *s3.Client
}

// New creates a new registry client with validated configuration
func New(cfg Config) (*Client, error) {
	slog.Info("Creating a new registry instance")

	// Validate configuration first
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create registry client: %w", err)
	}

	// Load AWS config with custom endpoint if provided
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Default region, can be made configurable
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided
	s3Options := []func(*s3.Options){}
	if cfg.S3Endpoint != "" {
		s3Options = append(s3Options, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true // Often needed for custom endpoints
		})
	}

	s3Client := s3.NewFromConfig(awsCfg, s3Options...)

	client := &Client{
		CFG: cfg,
		S3:  s3Client,
	}

	// Optional: Test the connection
	if err := client.testConnection(context.TODO()); err != nil {
		return nil, fmt.Errorf("failed to test registry connection: %w", err)
	}

	slog.Info("Registry client created successfully",
		"endpoint", cfg.S3Endpoint,
		"bucket", cfg.Bucket,
		"prefix", cfg.Prefix)

	return client, nil
}

// testConnection verifies that the client can communicate with S3
func (client *Client) testConnection(ctx context.Context) error {
	_, err := client.S3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(client.CFG.Bucket),
	})
	if err != nil {
		return fmt.Errorf("cannot access bucket %s: %w", client.CFG.Bucket, err)
	}
	return nil
}

// constructS3Key generates the S3 key path for a file
func (client *Client) constructS3Key(filePath, sha256 string) string {
	ext := path.Ext(filePath)
	fileName := fmt.Sprintf("%s%s", sha256, ext)
	return path.Join(client.CFG.Prefix, fileName)
}