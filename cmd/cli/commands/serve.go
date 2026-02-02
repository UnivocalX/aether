package commands

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/logging"
	"github.com/UnivocalX/aether/internal/web"
	"github.com/UnivocalX/aether/internal/registry"
)

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:           "serve",
	Short:         "Starts the aether api server",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          startServer,
}

func init() {
	// Storage
	ServeCmd.Flags().Int("port", 8080, "Port to run the server on")
	ServeCmd.Flags().String("s3endpoint", "", "S3 endpoint")
	ServeCmd.Flags().String("bucket", "", "S3 bucket.")
	ServeCmd.Flags().String("prefix", "aether", "S3 prefix.")

	// Database
	ServeCmd.Flags().String("db-endpoint", "localhost:5432", "Database port.")
	ServeCmd.Flags().String("db-user", "postgres", "")
	ServeCmd.Flags().String("db-password", "changeme", "Port to run the server on")
	ServeCmd.Flags().String("db-name", "postgres", "Database name.")
	ServeCmd.Flags().Bool("ssl", false, "Database SSL.")
	ServeCmd.Flags().Bool("production", false, "Run in production mode (enables JSON logging)")

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
	Log := logging.NewLog()

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
	viper.BindPFlag("server.port", ServeCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.production", ServeCmd.Flags().Lookup("production"))

	// Storage settings
	viper.BindPFlag("server.storage.s3endpoint", ServeCmd.Flags().Lookup("s3endpoint"))
	viper.BindPFlag("server.storage.bucket", ServeCmd.Flags().Lookup("bucket"))
	viper.BindPFlag("server.storage.prefix", ServeCmd.Flags().Lookup("prefix"))

	// Database settings
	viper.BindPFlag("server.database.endpoint", ServeCmd.Flags().Lookup("db-endpoint"))
	viper.BindPFlag("server.database.user", ServeCmd.Flags().Lookup("db-user"))
	viper.BindPFlag("server.database.password", ServeCmd.Flags().Lookup("db-password"))
	viper.BindPFlag("server.database.name", ServeCmd.Flags().Lookup("db-name"))
	viper.BindPFlag("server.database.ssl", ServeCmd.Flags().Lookup("ssl"))
}
