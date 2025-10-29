package registry

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"time"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/UnivocalX/aether/pkg/registry/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define sentinel errors
var (
	ErrAssetNotFound = errors.New("asset not found")
	ErrTagNotFound   = errors.New("tag not found")
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
		"datastore", fmt.Sprintf("%s:%d", cfg.Datastore.Endpoint.GetHost(), cfg.Datastore.Endpoint.GetPort()),
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

// IngressKey generates the ingress S3 key path
func (engine *Engine) IngressKey(sha256 string) string {
	return path.Join(engine.Config.Storage.Prefix, "ingress", sha256)
}

// CuratedKey generates the curated S3 key path
func (engine *Engine) CuratedKey(sha256 string) string {
	return path.Join(engine.Config.Storage.Prefix, "curated", sha256)
}

// PutURL generates a presigned URL for upload (default expiry)
func (engine *Engine) PutURL(ctx context.Context, sha256 string) (string, error) {
	return engine.PutURLExpire(ctx, sha256, engine.Config.Storage.TTL)
}

// PutURLExpire generates a presigned URL with custom expiry and checksum validation
func (engine *Engine) PutURLExpire(ctx context.Context, sha256 string, expire time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:            aws.String(engine.Config.Storage.Bucket),
		Key:               aws.String(engine.IngressKey(sha256)),
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
		Key:    aws.String(engine.CuratedKey(sha256)),
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

func (engine *Engine) CreateAssetRecord(ctx context.Context, sha256 string, display string) (*models.Asset, error) {
	slog.Info("Creating asset", "display", display, "checksum", sha256)

	asset := &models.Asset{
		Checksum: sha256,
		Display:  display,
	}

	if err := engine.DB.Create(asset).Error; err != nil {
		return nil, err
	}

	return asset, nil
}

func (engine *Engine) GetAssetRecord(ctx context.Context, sha256 string) (*models.Asset, error) {
    normalizedSha256 := models.NormalizeName(sha256)
    slog.Info("Getting asset", "checksum", normalizedSha256)

    var asset models.Asset
    err := engine.DB.Where("checksum = ?", normalizedSha256).First(&asset).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (engine *Engine) CreateTagRecord(ctx context.Context, name string) (*models.Tag, error) {
	slog.Info(fmt.Sprintf("Creating tag: %s", name))

	tag := &models.Tag{
		Name: name,
	}

	if err := engine.DB.Create(tag).Error; err != nil {
		return nil, err
	}

	return tag, nil
}

func (engine *Engine) GetTagRecord(ctx context.Context, name string) (*models.Tag, error) {
	normalizedName := models.NormalizeName(name)
	slog.Debug("Getting tag", "name", normalizedName)

	var tag models.Tag
	err := engine.DB.Where("name = ?", normalizedName).First(&tag).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAssetNotFound
	}

	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (engine *Engine) AssociateTagWithAsset(assetID uint, tagID uint) error {
	slog.Debug("Attempting to associate tag with asset", "AssetID", assetID, "tagID", tagID)
	
	var asset models.Asset
	if err := engine.DB.First(&asset, assetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAssetNotFound
		}
		return err
	}

	var tag models.Tag
	if err := engine.DB.First(&tag, tagID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTagNotFound
		}
		return err
	}

	if err := engine.DB.Model(&asset).Association("Tags").Append(&tag); err != nil {
		return err
	}
	return nil
}
