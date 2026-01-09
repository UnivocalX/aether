package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/UnivocalX/aether/pkg/registry"
	"gorm.io/gorm"
)

func (s *Service) CreateDataset(ctx context.Context, name string, description string) (*registry.DatasetVersion, error) {
	slog.Debug("attempting to create a new dataset", "name", name)

	var dsv *registry.DatasetVersion
	err := s.engine.DatabaseClient.Transaction(func(tx *gorm.DB) error {
		engine := s.engine.WithTx(tx)

		// create dataset
		ds, err := engine.CreateDatasetRecord(name, description)
		if err != nil {
			if IsUniqueConstraintError(err) {
				return fmt.Errorf("%w: %s", ErrDatasetAlreadyExists, name)
			}

			return fmt.Errorf("failed to create dataset: %w", err)
		}

		// create first version
		dsv, err = engine.CreateDatasetVersionRecord(ds.Name, description)
		if err != nil {
			return err
		}

		return nil
	})

	return dsv, err
}