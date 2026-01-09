package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/UnivocalX/aether/pkg/registry"
	"gorm.io/gorm"
)

const (
	DefaultLimit = 100
	MaxLimit     = 1000
	MinLimit     = 1
)

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

func (s *Service) AddTag(ctx context.Context, name string) (*registry.Tag, error) {
	slog.Debug("attempting to create tag")

	// Try to fetch existing tag
	tag, err := s.engine.GetTagRecord(name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		// Tag not found → create it
		tag, err = s.engine.SaveTagRecord(name)
		if err != nil {
			return nil, err
		}
	}

	return tag, nil
}

func (s *Service) GetTag(ctx context.Context, name string) (*registry.Tag, error) {
	slog.Debug("attempting to get tag", "name", name)

	tag, err := s.engine.GetTagRecord(name)

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

	assets, err := s.engine.GetTagRecordAssets(params.Name, int(params.Limit), int(params.Offset))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTagNotFound, params.Name)
		}

		return nil, fmt.Errorf("failed to get tag assets: %w", err)
	}

	return assets, nil
}
