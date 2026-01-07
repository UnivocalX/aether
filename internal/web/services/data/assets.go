package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CreateAssetParams struct {
	Checksum string
	Display  string
	Tags     []string
	Extra    map[string]any
}

// Service layer result - internal representation
type CreateAssetResult struct {
	Asset     *registry.Asset
	UploadURL *registry.PresignedUrl
	Err       error
}

func (s *Service) CreateAsset(
	ctx context.Context,
	params CreateAssetParams,
) *CreateAssetResult {
	slog.Debug("Attempting to create asset")
	result := &CreateAssetResult{}

	// Wrap create and tag association in a transaction
	err := s.engine.DatabaseClient.Transaction(func(tx *gorm.DB) error {
		engine := s.engine.WithTx(tx)

		// Create asset record
		asset, err := engine.CreateAssetRecord(params.Checksum, params.Display, params.Extra)
		if err != nil {
			return err // rollback
		}
		slog.Debug("Created new asset", "checksum", params.Checksum)

		// Associate tags
		if err := attachTagsToAsset(engine, asset, params.Tags); err != nil {
			return err // rollback
		}

		result.Asset = asset
		return nil // commit
	})

	// handle errors
	if err != nil {
		slog.Debug("Create asset transaction failed", "checksum", params.Checksum, "error", err.Error())

		// Check PostgreSQL-specific error code
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("asset already exist", "checksum", params.Checksum)
			result.Err = fmt.Errorf("%w: %s", ErrAssetAlreadyExists, params.Checksum)
			return result
		}

		slog.Debug("unexpected error occurred", "error", err)
		result.Err = err
		return result
	}

	// Generate put URL (outside transaction - S3 operation)
	url, err := s.engine.IngressURL(ctx, params.Checksum)
	if err != nil {
		result.Err = fmt.Errorf("failed generating presigned URL: %w", err)
		return result
	}

	result.UploadURL = url
	return result
}

// attachTagsToAsset fetches tags and associates them with an asset
func attachTagsToAsset(engine *registry.Engine, asset *registry.Asset, tagsNames []string) error {
	slog.Debug("attempting to attach tags to new asset", "total", len(tagsNames))
	if len(tagsNames) == 0 {
		return nil
	}

	// Normalize tag names for comparison
	normalized := make([]string, 0, len(tagsNames))
	for _, name := range tagsNames {
		if n := registry.NormalizeString(name); n != "" {
			normalized = append(normalized, n)
		}
	}

	// Fetch all tags
	tags, err := engine.GetTagsByNames(normalized)
	if err != nil {
		return err
	}

	// Check if all tags were found
	if err := validateAllTagsFound(normalized, tags); err != nil {
		return err
	}

	// Associate tags with asset
	if err := engine.AttachTags(asset, tags); err != nil {
		return err
	}

	return nil
}

// validateAllTagsFound checks if all requested tags were found and returns an error with missing tag names if not
func validateAllTagsFound(names []string, tags []*registry.Tag) error {
	if len(tags) == len(names) {
		return nil
	}

	// Build a map of found tags
	found := make(map[string]bool, len(tags))
	for _, tag := range tags {
		found[tag.Name] = true
	}

	// Find missing tags
	missing := make([]string, 0)
	for _, name := range names {
		if !found[name] {
			missing = append(missing, name)
		}
	}

	return fmt.Errorf("%w: %v", ErrTagNotFound, missing)
}

func (s *Service) GetAsset(ctx context.Context, checksum string) (*registry.Asset, error) {
	slog.Debug("attempting to get asset", "checksum", checksum)
	asset, err := s.engine.GetAssetRecord(checksum)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, checksum)
		}

		return nil, err
	}

	return asset, nil
}

func (s *Service) TagAsset(ctx context.Context, checksum string, tagName string) error {
	slog.Debug("attempting to tag asset", "tagName", tagName, "assetChecksum", checksum)

	asset, err := s.GetAsset(ctx, checksum)
	if err != nil {
		return err
	}

	tag, err := s.GetTag(ctx, tagName)
	if err != nil {
		return err
	}

	if err := s.engine.AttachTags(asset, []*registry.Tag{tag}); err != nil {
		return fmt.Errorf("failed to tag asset: %w", err)
	}

	return nil
}

func (s *Service) UntagAsset(ctx context.Context, checksum string, tagName string) error {
	slog.Debug("attempting to untag asset", "tagName", tagName, "assetChecksum", checksum)

	asset, err := s.GetAsset(ctx, checksum)
	if err != nil {
		return err
	}

	tag, err := s.GetTag(ctx, tagName)
	if err != nil {
		return err
	}

	if err := s.engine.DetachTags(asset, []*registry.Tag{tag}); err != nil {
		return fmt.Errorf("failed to untag asset: %w", err)
	}

	return nil
}

func (s *Service) GetAssetTags(ctx context.Context, checksum string) ([]*registry.Tag, error) {
	slog.Debug("attempting to get asset tags", "checksum", checksum)

	tags, err := s.engine.GetAssetRecordTags(checksum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, checksum)
		}

		return nil, err
	}

	return tags, nil
}

func (s *Service) GetAssetIngressUrl(ctx context.Context, checksum string) (*registry.PresignedUrl, error) {
	slog.Debug("attempting to get asset Presigned Url", "checksum", checksum)

	// Check asset status
	asset, err := s.engine.GetAssetRecord(checksum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, checksum)
		}

		return nil, fmt.Errorf("failed to get asset presigned url: %w", err)
	}

	// reuploading a ready asset is not allowed
	if asset.State == registry.StatusReady {
		slog.Warn("attempt to get presigned url for ready asset", "checksum", checksum)
		return nil, fmt.Errorf("%w: %s", ErrAssetIsReady, checksum)
	}

	return s.engine.IngressURL(ctx, checksum)
}

func (s *Service) ListAssets(ctx context.Context, opts ...registry.SearchAssetsOption) ([]*registry.Asset, error) {
	slog.Debug("attempting to list assets")
	return s.engine.ListAssetsRecords(opts...)
}

func (s *Service) CreateAssets(ctx context.Context, assets ...*registry.Asset) error {
	slog.Debug("attempting to create new assets", "total", len(assets))

	// Check if any of the records exist already
	checksums := make([]string, len(assets))
	for i, record := range assets {
		checksums[i] = record.Checksum
	}

	existingRecords, err := s.engine.ListAssetsRecords(registry.WithChecksums(checksums...))
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(existingRecords) > 0 {
		existingChecksums := make([]string, len(assets))
		for i, record := range existingRecords {
			existingChecksums[i] = record.Checksum
		}

		return ErrAssetsExists{Checksums: existingChecksums}
	}

	// Create records
	return s.engine.CreateAssetRecords(assets...)
}
