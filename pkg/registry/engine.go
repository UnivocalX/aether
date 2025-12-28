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
func (engine *Engine) PutURL(ctx context.Context, sha256 string) (*PresignedUrl, error) {
	return engine.PutUrlExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

func (engine *Engine) PutUrlExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignedUrl, error) {
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

	presignUrl := &PresignedUrl{
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
func (engine *Engine) GetURL(ctx context.Context, sha256 string) (*PresignedUrl, error) {
	return engine.GetUrlExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

// GetUrlExpire generates a presigned URL for download with custom expiry
func (engine *Engine) GetUrlExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignedUrl, error) {
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

	presignUrl := &PresignedUrl{
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
	slog.Debug("Getting asset", "checksum", sha256)
	normalizedSha256 := NormalizeString(sha256)

	var asset Asset
	err := engine.db(ctx).
		Where("checksum = ?", normalizedSha256).
		First(&asset).Error

	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (engine *Engine) GetAssetRecordTags(ctx context.Context, sha256 string) ([]*Tag, error) {
	slog.Debug("Getting asset tags", "checksum", sha256)

	asset, err := engine.GetAssetRecord(ctx, sha256)
	if err != nil {
		return nil, err
	}

	var tags []*Tag
	if err := engine.db(ctx).Model(asset).Association("Tags").Find(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

func (engine *Engine) CreateAssetRecord(
	ctx context.Context,
	sha256 string,
	display string,
	extra map[string]any,
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

func (engine *Engine) AttachTags(ctx context.Context, asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to attach tags to asset", "AssetID", asset.ID, "tagCount", len(tags))

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

func (engine *Engine) DetachTags(ctx context.Context, asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to detach tags from asset", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Append all tags at once
	if err := engine.db(ctx).Model(asset).Association("Tags").Delete(tags); err != nil {
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

func (engine *Engine) GetTagRecord(ctx context.Context, name string) (*Tag, error) {
	normalizedName := NormalizeString(name)
	slog.Debug("Getting tag", "name", name)

	var tag Tag
	err := engine.db(ctx).
		Where("name = ?", normalizedName).
		First(&tag).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (engine *Engine) GetTagRecordAssets(ctx context.Context, name string, limit int, offset int) ([]*Asset, error) {
    tag, err := engine.GetTagRecord(ctx, name)
    if err != nil {
        return nil, err
    }

    var assets []*Asset
    err = engine.db(ctx).
        Limit(limit).
        Offset(offset).
        Model(tag).
        Association("Assets").
        Find(&assets)
    
    if err != nil {
        return nil, err
    }

    return assets, nil
}

func (engine *Engine) GetTagRecordById(ctx context.Context, id uint) (*Tag, error) {
	slog.Debug("Getting tag", "id", id)

	var tag Tag
	if err := engine.db(ctx).First(&tag, id).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

// GetTagsByNames fetches multiple tags by their names in a single query
func (engine *Engine) GetTagsByNames(ctx context.Context, names []string) ([]*Tag, error) {
	slog.Debug("Getting tags", "total", len(names))
	if len(names) == 0 {
		return nil, nil
	}

	// Normalize all names
	slog.Debug("normalizing tags names")
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		if n := NormalizeString(name); n != "" {
			normalized = append(normalized, n)
		}
	}

	if len(normalized) == 0 {
		return nil, nil
	}

	var tags []*Tag
	err := engine.db(ctx).
		Where("name IN ?", normalized).
		Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (engine *Engine) ListAssetsRecords(ctx context.Context, opts ...SearchAssetsOption) ([]*Asset, error) {
	slog.Debug("Listing assets", "totalOptions", len(opts))

	query, err := NewSearchAssetsQuery(opts...)
	if err != nil {
		slog.Debug("failed to build query")
		return nil, err
	}
	slog.Debug("created new query", "query", query)

	db := engine.db(ctx).Where(&Asset{MimeType: query.MimeType, State: query.State})

	// IncludedTags: Filter assets that have ALL specified tags (AND logic)
	if len(query.IncludedTags) > 0 {
		db = db.Where("id IN (?)", // Filter main query to only IDs from subquery
			engine.db(ctx). // Create fresh DB connection for subquery
					Table("assets").                                             // Query the assets table
					Select("assets.id").                                         // Return only asset IDs
					Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id"). // Connect assets to junction table
					Joins("JOIN tags ON tags.id = asset_tags.tag_id").           // Connect junction table to tags
					Where("tags.name IN ?", query.IncludedTags).                 // Keep only rows with required tag names
					Group("assets.id").                                          // Group rows by asset to count tags
					Having("COUNT(*) = ?", len(query.IncludedTags)),             // Only keep assets with exact tag count (ALL tags present)
		)
	}

	// ExcludedTags: Filter out assets that have ANY of these tags
	if len(query.ExcludedTags) > 0 {
		db = db.Where("id NOT IN (?)", // Exclude IDs that match subquery
			engine.db(ctx). // Create fresh DB connection for subquery
					Table("assets").                                             // Query the assets table
					Select("assets.id").                                         // Return only asset IDs
					Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id"). // Connect assets to junction table
					Joins("JOIN tags ON tags.id = asset_tags.tag_id").           // Connect junction table to tags
					Where("tags.name IN ?", query.ExcludedTags),                 // Find assets with any excluded tag
		)
	}

	// Pagination
	db = db.Where("id > ?", query.Cursor)
	db = db.Limit(int(query.Limit))

	// Execute query with preloaded tags
	var assets []*Asset
	if err := db.Preload("Tags").Order("id ASC").Find(&assets).Error; err != nil {
		return nil, err
	}

	return assets, nil
}
