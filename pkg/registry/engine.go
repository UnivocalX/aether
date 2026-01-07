package registry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	DEFAULT_DATABASE_NAME = "postgres"
	DEFAULT_PRESIGN_TTL   = 15 * time.Minute
	DEFAULT_TIME_ZONE     = "UTC"
	DEFAULT_DATABASE      = "localhost:5432"
)

type Engine struct {
	// storage
	storage Endpoint
	bucket  string
	prefix  string

	// database
	database         Endpoint
	databaseUser     string
	databasePassword Secret
	databaseName     string
	databaseSslMode  bool

	// global
	timeZone string

	// clients
	S3Client       *s3.Client
	PresignClient  *s3.PresignClient
	DatabaseClient *gorm.DB
}

// New creates new core engine
func New(opts ...Option) (*Engine, error) {
	engine := &Engine{
		database:     DEFAULT_DATABASE,
		databaseName: DEFAULT_DATABASE_NAME,
		timeZone:     DEFAULT_TIME_ZONE,
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(engine); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create S3 Client
	if err := engine.createS3Client(); err != nil {
		return nil, err
	}

	// Create DB Client
	if err := engine.createDatabaseClient(); err != nil {
		return nil, err
	}

	return engine, nil
}

func (engine *Engine) dsn() DSN {
	sslMode := "disable"
	if engine.databaseSslMode {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		engine.database.GetHost("localhost"), engine.database.GetPort(5432),
		engine.databaseUser, engine.databasePassword.Value(), engine.databaseName, sslMode, engine.timeZone,
	)

	return DSN(dsn)
}

func (engine *Engine) createS3Client() error {
	// AWS Client
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if engine.storage != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(string(engine.storage))
			o.UsePathStyle = true
		})
	}

	engine.S3Client = s3.NewFromConfig(awsCfg, s3Opts...)
	engine.PresignClient = s3.NewPresignClient(engine.S3Client)
	return nil
}

func (engine *Engine) createDatabaseClient() error {
	dsn := engine.dsn()
	slog.Debug("Database connection details", "dsn", dsn)

	gormLogger := slogGorm.New() // use slog.Default() by default
	db, err := gorm.Open(
		postgres.Open(dsn.Value()),
		&gorm.Config{Logger: gormLogger, CreateBatchSize: 1000})

	if err != nil {
		return err
	}
	engine.DatabaseClient = db

	// Run migrations
	if err := engine.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
