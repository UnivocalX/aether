package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	DefaultLimit = 100
	MaxLimit     = 1000
	MinLimit     = 1
)

type CreateTagParams struct {
	Name string
}

type CreateTagResult struct {
	Tag *registry.Tag
	Err error
}

type GetTagAssetsParams struct {
	Name   string
	Offset uint
	Limit  uint
}

func (p *GetTagAssetsParams) Validate() error {
	// Validate name
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("tag name cannot be empty")
	}

	// Set default limit if not provided
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}

	// Enforce minimum limit
	if p.Limit < MinLimit {
		p.Limit = MinLimit
	}

	// Enforce maximum limit
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}

	// Offset is already uint, so it can't be negative
	// No validation needed for offset (can be 0 or any positive number)

	return nil
}

type GetTagAssetsResult struct {
	Assets []*registry.Asset
}

func (s *Service) CreateTag(ctx context.Context, params CreateTagParams) *CreateTagResult {
	slog.Debug("attempting to create tag")
	result := &CreateTagResult{}

	db := s.engine.WithDatabase(ctx)
	tag, err := registry.CreateTagRecord(db, params.Name)
	if err != nil {
		// Check PostgreSQL-specific error code
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slog.Error("tag already exist", "name", params.Name)
			result.Err = fmt.Errorf("%w: %s", ErrTagAlreadyExists, params.Name)
			return result
		}

		result.Err = err
		return result
	}

	result.Tag = tag
	return result
}

func (s *Service) GetTag(ctx context.Context, name string) (*registry.Tag, error) {
	slog.Debug("attempting to get tag", "name", name)

	db := s.engine.WithDatabase(ctx)
	tag, err := registry.GetTagRecord(db, name)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTagNotFound, name)
		}

		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (s *Service) GetTagAssets(ctx context.Context, params GetTagAssetsParams) ([]*registry.Asset, error) {
	// Validate and apply defaults
	if err := params.Validate(); err != nil {
		return nil, err
	}

	slog.Debug("attempting to get tag assets",
		"name", params.Name,
		"limit", params.Limit,
		"offset", params.Offset,
	)

	db := s.engine.WithDatabase(ctx)
	assets, err := registry.GetTagRecordAssets(db, params.Name, int(params.Limit), int(params.Offset))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTagNotFound, params.Name)
		}

		return nil, fmt.Errorf("failed to get tag assets: %w", err)
	}

	return assets, nil
}
