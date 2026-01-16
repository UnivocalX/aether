package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/logging"
	"github.com/UnivocalX/aether/internal/web"
	"github.com/UnivocalX/aether/pkg/registry"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:           "serve",
	Short:         "Starts the aether api server",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          startServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Storage
	serveCmd.Flags().Int("port", 8080, "Port to run the server on")
	serveCmd.Flags().String("s3endpoint", "", "S3 endpoint")
	serveCmd.Flags().String("bucket", "", "S3 bucket.")
	serveCmd.Flags().String("prefix", "aether", "S3 prefix.")

	// Database
	serveCmd.Flags().String("db-endpoint", "localhost:5432", "Database port.")
	serveCmd.Flags().String("db-user", "postgres", "")
	serveCmd.Flags().String("db-password", "changeme", "Port to run the server on")
	serveCmd.Flags().String("db-name", "postgres", "Database name.")
	serveCmd.Flags().Bool("ssl", false, "Database SSL.")
	serveCmd.Flags().Bool("production", false, "Run in production mode (enables JSON logging)")

	bindServeFlags()
}

func startServer(cmd *cobra.Command, args []string) error {
	// Setup server logging
	slog.Debug("changing logging mode to server mode")
	prod := viper.GetBool("server.production")
	if err := updateLogging(prod); err != nil {
		return err
	}

	// Create Core Engine(registry engine)
	slog.Info("Initializing registry engine")
	engine, err := initRegistry()
	if err != nil {
		return err
	}

	// Run server
	port := viper.GetString("server.port")
	server := web.NewServer(prod, engine)
	return server.Run(port)
}

func updateLogging(prod bool) error {
	Log.SetMode(logging.ServerMode)
	if !prod {
		Log.EnableColor()
	}
	Log.Apply()
	return nil
}

func initRegistry() (*registry.Engine, error) {
	opts := getRegistryOptions()
	engine, err := registry.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registry: %w", err)
	}
	return engine, nil
}

func getRegistryOptions() []registry.Option {
	opts := []registry.Option{
		registry.WithBucket(viper.GetString("server.storage.bucket")),
	}

	addIfSet := func(key string, optFunc func(string) registry.Option) {
		if val := viper.GetString(key); val != "" {
			opts = append(opts, optFunc(val))
		}
	}

	addIfSet("server.storage.s3endpoint", registry.WithStorageEndpoint)
	addIfSet("server.storage.prefix", registry.WithBucketPrefix)
	addIfSet("server.database.endpoint", registry.WithDatabaseEndpoint)
	addIfSet("server.database.user", registry.WithDatabaseUser)
	addIfSet("server.database.password", registry.WithDatabasePassword)
	addIfSet("server.database.name", registry.WithDatabaseName)

	if viper.GetBool("server.database.ssl") {
		opts = append(opts, registry.WithSslMode())
	}

	return opts
}

func bindServeFlags() {
	// Server settings
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.production", serveCmd.Flags().Lookup("production"))

	// Storage settings
	viper.BindPFlag("server.storage.s3endpoint", serveCmd.Flags().Lookup("s3endpoint"))
	viper.BindPFlag("server.storage.bucket", serveCmd.Flags().Lookup("bucket"))
	viper.BindPFlag("server.storage.prefix", serveCmd.Flags().Lookup("prefix"))

	// Database settings
	viper.BindPFlag("server.database.endpoint", serveCmd.Flags().Lookup("db-endpoint"))
	viper.BindPFlag("server.database.user", serveCmd.Flags().Lookup("db-user"))
	viper.BindPFlag("server.database.password", serveCmd.Flags().Lookup("db-password"))
	viper.BindPFlag("server.database.name", serveCmd.Flags().Lookup("db-name"))
	viper.BindPFlag("server.database.ssl", serveCmd.Flags().Lookup("ssl"))
}
