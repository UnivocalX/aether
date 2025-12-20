package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateAssetParams struct {
	SHA256  string
	Display string
	Tags    []uint
	Extra   map[string]interface{}
}

// Service layer result - internal representation
type CreateAssetResult struct {
	Asset     *registry.Asset
	UploadURL *registry.PresignUrl
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

// handleCreateAssetTags fetches tags and associates them with an asset
func (s *Service) handleCreateAssetTags(ctx context.Context, asset *registry.Asset, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// Fetch all tags
	tags, err := s.getTagsByIDs(ctx, tagIDs)
	if err != nil {
		return err
	}

	// Associate tags with asset
	if err := s.registry.UpdateAssetTags(ctx, asset, tags); err != nil {
		return fmt.Errorf("failed associating tags with asset: %w", err)
	}

	return nil
}