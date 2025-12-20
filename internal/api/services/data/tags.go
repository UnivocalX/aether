package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/jackc/pgx/v5/pgconn"
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

// getTagsByIDs fetches tags by their IDs
func (s *Service) getTagsByIDs(ctx context.Context, tagIDs []uint) ([]*registry.Tag, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}

	tags := make([]*registry.Tag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		tag, err := s.registry.GetTagRecord(ctx, tagID)
		if err != nil {
			return nil, fmt.Errorf("failed getting tag %d: %w", tagID, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
