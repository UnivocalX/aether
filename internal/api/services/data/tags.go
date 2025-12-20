package data

import (
	"context"
	"fmt"

	"github.com/UnivocalX/aether/pkg/registry"
)

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