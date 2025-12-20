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

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	DEFAULT_DATABASE_NAME = "postgres"
	DEFAULT_PRESIGN_TTL   = 15 * time.Minute
	DEFAULT_TIME_ZONE     = "UTC"
	DEFAULT_DATABASE      = "localhost:5432"
)

type Engine struct {
	// storage
	storage Endpoint
	bucket  string
	prefix  string

	// database
	database         Endpoint
	databaseUser     string
	databasePassword Secret
	databaseName     string
	databaseSslMode  bool

	// global
	timeZone string

	// clients
	S3Client       *s3.Client
	PresignClient  *s3.PresignClient
	DatabaseClient *gorm.DB
}

// New creates new core engine
func New(opts ...Option) (*Engine, error) {
	engine := &Engine{
		database:     DEFAULT_DATABASE,
		databaseName: DEFAULT_DATABASE_NAME,
		timeZone:     DEFAULT_TIME_ZONE,
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(engine); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create S3 Client
	if err := engine.createS3Client(); err != nil {
		return nil, err
	}

	// Create DB Client
	if err := engine.createDatabaseClient(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (engine *Engine) dsn() DSN {
	sslMode := "disable"
	if engine.databaseSslMode {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		engine.database.GetHost("localhost"), engine.database.GetPort(5432),
		engine.databaseUser, engine.databasePassword.Value(), engine.databaseName, sslMode, engine.timeZone,
	)

	return DSN(dsn)
}

func (engine *Engine) createS3Client() error {
	// AWS Client
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if engine.storage != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(string(engine.storage))
			o.UsePathStyle = true
		})
	}

	engine.S3Client = s3.NewFromConfig(awsCfg, s3Opts...)
	engine.PresignClient = s3.NewPresignClient(engine.S3Client)
	return nil
}

func (engine *Engine) createDatabaseClient() error {
	dsn := engine.dsn()
	slog.Debug("Database connection details", "dsn", dsn)

	gormLogger := slogGorm.New() // use slog.Default() by default
	db, err := gorm.Open(postgres.Open(dsn.Value()), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return err
	}
	engine.DatabaseClient = db

	// Run migrations
	if err := engine.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

// Ping verifies S3 connection
func (engine *Engine) PingStorage(ctx context.Context) error {
	_, err := engine.S3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(engine.bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket access: %w", err)
	}
	return nil
}

// IngressKey generates the ingress S3 key path
func (engine *Engine) IngressKey(sha256 string) string {
	return path.Join(engine.prefix, "ingress", sha256)
}

// CuratedKey generates the curated S3 key path
func (engine *Engine) CuratedKey(sha256 string) string {
	return path.Join(engine.prefix, "curated", sha256)
}

// PutURL generates a presigned URL for upload (default expiry)
func (engine *Engine) PutURL(ctx context.Context, sha256 string) (*PresignUrl, error) {
	return engine.PutURLExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

func (engine *Engine) PutURLExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignUrl, error) {
	key := engine.IngressKey(sha256)

	input := &s3.PutObjectInput{
		Bucket:            aws.String(engine.bucket),
		Key:               aws.String(key),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ChecksumSHA256:    aws.String(sha256),
	}

	res, err := engine.PresignClient.PresignPutObject(ctx, input,
		s3.WithPresignExpires(expire),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate presign put url: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(expire)

	presignUrl := &PresignUrl{
		URL:       Secret(res.URL),
		ExpiresAt: expiresAt,
		ExpiresIn: expire,
		Checksum:  sha256,
		Key:       key,
		Operation: "put",
		Bucket:    engine.bucket,
	}

	slog.Debug("Generated PUT URL with checksum validation",
		"sha256", sha256,
		"expire", expire,
		"expiresAt", expiresAt)

	return presignUrl, nil
}

// GetURL generates a presigned URL for download (default expiry)
func (engine *Engine) GetURL(ctx context.Context, sha256 string) (*PresignUrl, error) {
	return engine.GetURLExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

// GetURLExpire generates a presigned URL for download with custom expiry
func (engine *Engine) GetURLExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignUrl, error) {
	key := engine.CuratedKey(sha256)

	input := &s3.GetObjectInput{
		Bucket: aws.String(engine.bucket),
		Key:    aws.String(key),
	}

	res, err := engine.PresignClient.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expire),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presign get url: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(expire)

	presignUrl := &PresignUrl{
		URL:       Secret(res.URL),
		ExpiresAt: expiresAt,
		ExpiresIn: expire,
		Checksum:  sha256,
		Key:       key,
		Operation: "get",
		Bucket:    engine.bucket,
	}

	slog.Debug("Generated GET URL",
		"sha256", sha256,
		"expire", expire,
		"expiresAt", expiresAt)

	return presignUrl, nil
}

func (engine *Engine) GetAssetRecord(ctx context.Context, sha256 string) (*Asset, error) {
	normalizedSha256 := NormalizeString(sha256)
	slog.Debug("Getting asset", "checksum", normalizedSha256)

	var asset Asset
	err := engine.db(ctx).
		Where("checksum = ?", normalizedSha256).
		First(&asset).Error

	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (engine *Engine) CreateAssetRecord(
	ctx context.Context,
	sha256 string,
	display string,
	extra map[string]interface{},
) (*Asset, error) {
	slog.Debug("Creating asset record", "display", display, "checksum", sha256)

	asset := &Asset{
		Checksum: sha256,
		Display:  display,
	}

	// check for extra values
	if extra != nil {
		if err := asset.SetExtra(extra); err != nil {
			return nil, err
		}
	}

	if err := engine.db(ctx).Create(asset).Error; err != nil {
		return nil, err
	}

	return asset, nil
}

func (engine *Engine) UpdateAssetRecord(ctx context.Context, asset *Asset) error {
	result := engine.db(ctx).
		Model(asset).
		Updates(asset)

	if result.Error != nil {
		return result.Error
	}

	// Reload to get fresh timestamps, but reuse the same pointer
	return engine.db(ctx).
		First(asset, asset.ID).Error
}

func (engine *Engine) UpdateAssetTags(ctx context.Context, asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to update asset tags", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Append all tags at once
	if err := engine.db(ctx).Model(asset).Association("Tags").Append(tags); err != nil {
		return err
	}

	// Single reload for all tags
	if err := engine.db(ctx).Preload("Tags").First(asset, asset.ID).Error; err != nil {
		return err
	}

	return nil
}

func (engine *Engine) CreateTagRecord(ctx context.Context, name string) (*Tag, error) {
	slog.Debug(fmt.Sprintf("Creating tag: %s", name))

	tag := &Tag{
		Name: name,
	}

	if err := engine.db(ctx).Create(tag).Error; err != nil {
		return nil, err
	}

	return tag, nil
}

func (engine *Engine) GetTagRecord(ctx context.Context, id uint) (*Tag, error) {
	slog.Debug("Getting tag", "ID", id)

	var tag Tag
	if err := engine.db(ctx).First(&tag, id).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

func (engine *Engine) ListTags(ctx context.Context, cursor uint, limit int) ([]*Tag, uint, bool, error) {
	if limit <= 0 {
		limit = 50
	}

	fetchLimit := limit + 1
	var tags []*Tag

	query := engine.db(ctx).Limit(fetchLimit)

	if cursor > 0 {
		query = query.Where("id > ?", cursor)
	}

	if err := query.Order("id ASC").Find(&tags).Error; err != nil {
		return nil, 0, false, err
	}

	hasMore := len(tags) > limit
	if hasMore {
		tags = tags[:limit]
	}

	nextCursor := uint(0)
	if len(tags) > 0 {
		nextCursor = tags[len(tags)-1].ID
	}

	return tags, nextCursor, hasMore, nil
}

func (engine *Engine) ListAssets(ctx context.Context, cfg *SearchAssetsConfig) ([]*Asset, uint, bool, error) {
	// Use the constructor function and validate
	if cfg == nil {
		defaultOpts := NewSearchAssetsOptions()
		cfg = &defaultOpts
	}

	// Validate and normalize options using the struct methods
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, 0, false, fmt.Errorf("invalid search options: %w", err)
	}

	// Build query
	query := engine.buildSearchAssetQuery(ctx, cfg)

	// Preload relationships
	query = query.Preload("Tags").Preload("Peers").Preload("DatasetVersions")

	// Get total count for the current filters
	var totalCount int64
	countQuery := engine.buildSearchAssetQuery(ctx, cfg).Model(&Asset{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, false, fmt.Errorf("failed to count assets: %w", err)
	}

	// Apply pagination (fetch +1 to check for more results)
	// Convert uint to int for GORM Limit()
	limit := int(cfg.Limit)
	query = query.Order("id ASC").Limit(limit + 1)

	// Execute query
	var assets []*Asset
	if err := query.Find(&assets).Error; err != nil {
		return nil, 0, false, fmt.Errorf("failed to fetch assets: %w", err)
	}

	// Determine pagination info
	hasMore := len(assets) > limit
	if hasMore {
		assets = assets[:limit] // Remove the extra item
	}

	nextCursor := uint(0)
	if len(assets) > 0 {
		nextCursor = assets[len(assets)-1].ID
	}

	return assets, nextCursor, hasMore, nil
}

func (engine *Engine) buildSearchAssetQuery(ctx context.Context, cfg *SearchAssetsConfig) *gorm.DB {
	query := engine.db(ctx).Model(&Asset{})

	if cfg.MimeType != "" {
		query = query.Where("mime_type = ?", cfg.MimeType)
	}

	if cfg.State != "" {
		query = query.Where("state = ?", cfg.State)
	}

	if cfg.Cursor > 0 {
		query = query.Where("id > ?", cfg.Cursor)
	}

	// Safer included tags handling
	for _, tag := range cfg.IncludedTags {
		query = query.Where("EXISTS (?)",
			engine.db(ctx).Select("1").
				Table("asset_tags").
				Joins("JOIN tags ON tags.id = asset_tags.tag_id").
				Where("asset_tags.asset_id = assets.id").
				Where("tags.name = ?", tag),
		)
	}

	// Handle excluded tags
	if len(cfg.ExcludedTags) > 0 {
		query = query.Where("NOT EXISTS (?)",
			engine.db(ctx).Select("1").
				Table("asset_tags").
				Joins("JOIN tags ON tags.id = asset_tags.tag_id").
				Where("asset_tags.asset_id = assets.id").
				Where("tags.name IN ?", cfg.ExcludedTags),
		)
	}

	return query
}
