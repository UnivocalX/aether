package registry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	CFG Config
	S3  *s3.Client
}

// UploadResult holds the result of a single file upload
type UploadResult struct {
	FilePath string
	Key      string
	Err      error
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

// uploadFile uploads a single file to S3 and returns the S3 key and any error
func (client *Client) uploadFile(ctx context.Context, uploader *manager.Uploader, filePath string) (string, error) {
	sha256, err := CalculateSHA256(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", filePath, err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	key := client.constructS3Key(filePath, sha256)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:         aws.String(client.CFG.Bucket),
		Key:            aws.String(key),
		ChecksumSHA256: aws.String(sha256),
		Body:           file,
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload %s to S3: %w", filePath, err)
	}

	return key, nil
}

// upload uploads multiple files to S3 and returns results for each file
func (client *Client) uploadFiles(filePaths []string) ([]string, error) {
	slog.Debug("Uploading files", slog.Int("TotalFiles", len(filePaths)))

	// Create upload manager for optimized multipart uploads
	uploader := manager.NewUploader(client.S3)
	ctx := context.Background()

	// Create progress tracker if Rich mode is enabled
	var tracker *ProgressTracker
	if client.CFG.Rich {
		tracker = NewProgressTracker(len(filePaths))
		defer tracker.Finish()
	}

	var keys []string
	for _, filePath := range filePaths {
		key, err := client.uploadFile(ctx, uploader, filePath)

		if err != nil {
			return keys, err
		}

		keys = append(keys, key)

		// Update progress bar if enabled
		if tracker != nil {
			tracker.Increment()
		}
	}

	return keys, nil
}

func (client *Client) Load(pathPattern string) ([]string, error) {
	slog.Info("Loading files", slog.String("Pattern", pathPattern))
	var loaded []string

	// Gather Files
	filePaths, err := filepath.Glob(pathPattern)
	if err != nil {
		return loaded, err
	}

	totalFindings := len(filePaths)
	if totalFindings <= 0 {
		return loaded, fmt.Errorf("failed to find fitting files in %s", pathPattern)
	}
	slog.Info("Found fitting file(s).", slog.Int("TotalFiles", totalFindings))

	// Upload Files
	keys, err := client.uploadFiles(filePaths)
	if err != nil {
		return loaded, err
	}

	loaded = keys
	return loaded, nil
}
