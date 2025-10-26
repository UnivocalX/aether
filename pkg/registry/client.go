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

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Engine struct {
	Config *Config
	S3     *s3.Client
	Pre    *s3.PresignClient
	DB     *gorm.DB
}

// New creates new core engine
func New(cfg *Config) (*Engine, error) {
	slog.Debug("Creating engine")

	if err := cfg.Normalize(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// AWS Client
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if cfg.Storage.S3Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Storage.S3Endpoint)
			o.UsePathStyle = true
		})
	}

	s3Client := s3.NewFromConfig(awsCfg, s3Opts...)
	preClient := s3.NewPresignClient(s3Client)

	// DB Client
	dsn := cfg.Datastore.DSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Create engine
	engine := &Engine{
		Config: cfg,
		S3:     s3Client,
		Pre:    preClient,
		DB:     db,
	}

	// Use background context for connection test during initialization
	if err := engine.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	slog.Debug("Engine ready",
		"s3endpoint", cfg.Storage.S3Endpoint,
		"bucket", cfg.Storage.Bucket,
		"prefix", cfg.Storage.Prefix,
		"datastore", fmt.Sprintf("%s:%d", cfg.Datastore.Endpoint.GetHost(), cfg.Datastore.Endpoint.GetPort),
	)

	return engine, nil
}

// Ping verifies S3 connection
func (engine *Engine) Ping(ctx context.Context) error {
	_, err := engine.S3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(engine.Config.Storage.Bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket access: %w", err)
	}
	return nil
}

// Key generates the S3 key path
func (engine *Engine) Key(sha256 string) string {
	return path.Join(engine.Config.Storage.Prefix, sha256)
}

// PutURL generates a presigned URL for upload (default expiry)
func (engine *Engine) PutURL(ctx context.Context, sha256 string) (string, error) {
	return engine.PutURLExpire(ctx, sha256, engine.Config.Storage.TTL)
}

// PutURLExpire generates a presigned URL with custom expiry and checksum validation
func (engine *Engine) PutURLExpire(ctx context.Context, sha256 string, expire time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:            aws.String(engine.Config.Storage.Bucket),
		Key:               aws.String(engine.Key(sha256)),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256, // Enforce checksum validation
		ChecksumSHA256:    aws.String(sha256),            // Base64-encoded checksum
	}

	res, err := engine.Pre.PresignPutObject(ctx, input,
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
func (engine *Engine) GetURL(ctx context.Context, sha256 string) (string, error) {
	return engine.GetURLExpire(ctx, sha256, engine.Config.Storage.TTL)
}

// GetURLExpire generates a presigned URL for download with custom expiry
func (engine *Engine) GetURLExpire(ctx context.Context, sha256 string, expire time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(engine.Config.Storage.Bucket),
		Key:    aws.String(engine.Key(sha256)),
	}

	res, err := engine.Pre.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expire),
	)
	if err != nil {
		return "", fmt.Errorf("presign get: %w", err)
	}

	slog.Debug("Generated GET URL", "sha256", sha256, "expire", expire)
	return res.URL, nil
}
