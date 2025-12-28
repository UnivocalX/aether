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

type CreateTagParams struct {
	Name string
}

type CreateTagResult struct {
	Tag *registry.Tag
	Err error
}

func (s *Service) CreateTag(ctx context.Context, params CreateTagParams) *CreateTagResult {
	slog.Debug("attempting to create tag")
	result := &CreateTagResult{}

	tag, err := s.registry.CreateTagRecord(ctx, params.Name)
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
	tag, err := s.registry.GetTagRecord(ctx, name)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %s", ErrTagNotFound, name)
		}

		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}