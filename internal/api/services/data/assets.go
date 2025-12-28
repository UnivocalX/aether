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
	SHA256  string
	Display string
	Tags    []string
	Extra   map[string]interface{}
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
	err := s.registry.Transaction(ctx, func(txCtx context.Context) error {
		// Create asset record
		asset, err := s.registry.CreateAssetRecord(txCtx, params.SHA256, params.Display, params.Extra)
		if err != nil {
			return err // rollback
		}
		slog.Debug("Created new asset", "checksum", params.SHA256)

		// Associate tags
		if err := s.handleCreateAssetTags(txCtx, asset, params.Tags); err != nil {
			return err // rollback
		}

		result.Asset = asset
		return nil // commit
	})

	// handle errors
	if err != nil {
		slog.Debug("Create asset transaction failed", "sha256", params.SHA256, "error", err.Error())

		// Check PostgreSQL-specific error code
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("asset already exist", "sha256", params.SHA256)
			result.Err = fmt.Errorf("%w: %s", ErrAssetAlreadyExists, params.SHA256)
			return result
		}

		slog.Debug("unexpected error occurred", "error", err)
		result.Err = err
		return result
	}

	// Generate put URL (outside transaction - S3 operation)
	url, err := s.registry.PutURL(ctx, params.SHA256)
	if err != nil {
		result.Err = fmt.Errorf("failed generating presigned URL: %w", err)
		return result
	}

	result.UploadURL = url
	return result
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

// handleCreateAssetTags fetches tags and associates them with an asset
func (s *Service) handleCreateAssetTags(ctx context.Context, asset *registry.Asset, tagsNames []string) error {
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
	tags, err := s.registry.GetTagsByNames(ctx, tagsNames)
	if err != nil {
		return err
	}

	// Check if all tags were found
	if err := validateAllTagsFound(normalized, tags); err != nil {
		return err
	}

	// Associate tags with asset
	if err := s.registry.AttachTags(ctx, asset, tags); err != nil {
		return fmt.Errorf("failed associating tags with asset: %w", err)
	}

	return nil
}

func (s *Service) GetAsset(ctx context.Context, sha256 string) (*registry.Asset, error) {
	slog.Debug("attempting to get asset", "sha256", sha256)
	asset, err := s.registry.GetAssetRecord(ctx, sha256)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, sha256)
		}

		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	return asset, nil
}

func (s *Service) TagAsset(ctx context.Context, sha256 string, tagName string) error {
	slog.Debug("attempting to tag asset", "tagName", tagName, "assetSha256", sha256)

	asset, err := s.GetAsset(ctx, sha256)
	if err != nil {
		return err
	}

	tag, err := s.GetTag(ctx, tagName)
	if err != nil {
		return err
	}

	if err := s.registry.AttachTags(ctx, asset, []*registry.Tag{tag}); err != nil {
		return fmt.Errorf("failed to tag asset: %w", err)
	}

	return nil
}

func (s *Service) UntagAsset(ctx context.Context, sha256 string, tagName string) error {
	slog.Debug("attempting to untag asset", "tagName", tagName, "assetSha256", sha256)

	asset, err := s.GetAsset(ctx, sha256)
	if err != nil {
		return err
	}

	tag, err := s.GetTag(ctx, tagName)
	if err != nil {
		return err
	}

	if err := s.registry.DetachTags(ctx, asset, []*registry.Tag{tag}); err != nil {
		return fmt.Errorf("failed to untag asset: %w", err)
	}

	return nil
}

func (s *Service) GetAssetTags(ctx context.Context, sha256 string) ([]*registry.Tag, error) {
	slog.Debug("attempting to get asset tags", "sha256", sha256)
	tags, err := s.registry.GetAssetRecordTags(ctx, sha256)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, sha256)
		}

		return nil, fmt.Errorf("failed to get asset tags: %w", err)
	}

	return tags, nil
}

func (s *Service) GetAssetPresignedUrl(ctx context.Context, sha256 string) (*registry.PresignedUrl, error) {
	slog.Debug("attempting to get asset Presigned Url", "sha256", sha256)

	// Check asset status
	asset, err := s.registry.GetAssetRecord(ctx, sha256)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrAssetNotFound, sha256)
		}

		return nil, fmt.Errorf("failed to get asset presigned url: %w", err)
	}

	// reuploading a ready asset is not allowed
	if asset.State == registry.StatusReady {
		slog.Warn("attempt to get presigned url for ready asset", "Sha256", sha256)
		return nil, fmt.Errorf("%w: %s", ErrAssetIsReady, sha256)
	}

	return s.registry.PutURL(ctx, sha256)
}

func (s *Service) ListAssets(ctx context.Context, opts ...registry.SearchAssetsOption) ([]*registry.Asset, error) {
	slog.Debug("attempting to list assets")
	return s.registry.ListAssetsRecords(ctx, opts...)
}
