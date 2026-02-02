package registry

import (
	"fmt"
	"log/slog"
)

// Migrate runs database migrations
func (engine *Engine) Migrate() error {
	slog.Info("Running database migrations")

	// Step 1: Handle custom PostgreSQL types (AutoMigrate can't do this)
	if err := engine.createCustomTypes(); err != nil {
		return fmt.Errorf("failed to create custom types: %w", err)
	}

	// Step 2: Run AutoMigrate on all models
	if err := engine.autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	slog.Info("Database migrations completed successfully")
	return nil
}

// createCustomTypes creates PostgreSQL custom types that AutoMigrate can't handle
func (engine *Engine) createCustomTypes() error {
	// Create status enum type
	// Using DO $$ block makes it idempotent (won't fail if already exists)
	err := engine.DatabaseClient.Exec(`
		DO $$ BEGIN
			CREATE TYPE status AS ENUM ('pending', 'rejected', 'ready', 'deleted');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;
	`).Error

	if err != nil {
		return fmt.Errorf("failed to create status enum: %w", err)
	}

	// Add more custom types here as needed in the future:
	// e.g., CREATE TYPE priority AS ENUM ('low', 'medium', 'high');

	return nil
}

// autoMigrate runs GORM AutoMigrate on all models
func (engine *Engine) autoMigrate() error {
	// Register all your models here
	// Add new models to this list as your project grows
	return engine.DatabaseClient.AutoMigrate(
		// Core models
		&Asset{},
		&Tag{},
		&Dataset{},
		&DatasetVersion{},
		&Peer{},
	)
}
