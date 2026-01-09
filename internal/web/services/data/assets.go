package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"gorm.io/gorm"
)

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

func (s *Service) CreateAssets(ctx context.Context, assets ...*registry.Asset) ([]*registry.PresignedUrl, error) {
	slog.Debug("attempting to create new assets", "total", len(assets))

	// Try to create
	if err := s.engine.CreateAssetRecords(assets...); err != nil {
		// If duplicate error → fetch existing records and return them
		if IsUniqueConstraintError(err) {
			checksums := make([]string, len(assets))
			for i, a := range assets {
				checksums[i] = a.Checksum
			}

			existing, listErr := s.engine.ListAssetsRecords(registry.WithChecksums(checksums...))
			if listErr != nil {
				return nil, fmt.Errorf("failed to fetch existing assets after duplicate: %w", listErr)
			}

			return nil, AssetsExistsError{Checksums: registry.AssetsToChecksums(existing...)}
		}
		// other errors
		return nil, err
	}

	// Generate ingress urls
	checksums := registry.AssetsToChecksums(assets...)
	urls, err := s.GenerateIngressUrls(ctx, checksums...)
	if err != nil {
		return nil, err
	}

	return urls, err
}

func (s *Service) GenerateIngressUrls(ctx context.Context, checksums ...*string) ([]*registry.PresignedUrl, error) {
	slog.Debug("attempting to generate ingress urls", "total", len(checksums))
	ingress := make([]*registry.PresignedUrl, len(checksums))

	for i, c := range checksums {
		url, err := s.engine.IngressURL(ctx, *c)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrCantGeneratePresignedUrl, err)
		}
		ingress[i] = url
	}

	return ingress, nil
}
