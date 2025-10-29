package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/api"
	"github.com/UnivocalX/aether/pkg/registry"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the aether api server",
	RunE:  startServer,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Storage
	serveCmd.Flags().Int("port", 8080, "Port to run the server on")
	serveCmd.Flags().String("s3endpoint", "", "S3 endpoint")
	serveCmd.Flags().String("bucket", "", "S3 bucket.")
	serveCmd.Flags().String("prefix", "aether", "S3 prefix.")

	// Datastore
	serveCmd.Flags().String("db-endpoint", "localhost:5432", "Database port.")
	serveCmd.Flags().String("db-user", "postgres", "")
	serveCmd.Flags().String("db-password", "changeme", "Port to run the server on")
	serveCmd.Flags().String("db-name", "postgres", "Database name.")
	serveCmd.Flags().Bool("ssl", false, "Database SSL.")
}

func LoadEngineConfig() (*registry.Config, error) {
	cfg := registry.NewConfig()

	// Storage
	cfg.Storage.S3Endpoint = viper.GetString("s3endpoint")
	cfg.Storage.Bucket = viper.GetString("bucket")
	cfg.Storage.Prefix = viper.GetString("Prefix")

	// Datastore
	cfg.Datastore.Endpoint = registry.Endpoint(viper.GetString("db-endpoint"))
	cfg.Datastore.User = viper.GetString("db-user")
	cfg.Datastore.Password = viper.GetString("db-password")
	cfg.Datastore.Name = viper.GetString("db-name")
	cfg.Datastore.SSL = viper.GetBool("ssl")

	slog.Info("EngineConfig,",
		"Storage", cfg.Storage,
		"Datastore", cfg.Datastore,
	)

	return cfg, nil
}

func startServer(cmd *cobra.Command, args []string) error {
	slog.Info("Starting server")

	// Get Config
	port := viper.GetString("port")
	prod := viper.GetBool("production")

	cfg, err := LoadEngineConfig()
	if err != nil {
		return err
	}

	// Create engine
	engine, err := registry.New(cfg)
	if err != nil {
		slog.Error("Failed to create registry engine.")
		return err
	}

	// Create API
	router := api.New(engine, prod)

	// Run Server
	slog.Info("Serving:", "port", port, "mode", prod)
	if err := router.Run(":" + port); err != nil {
		slog.Error("Failed to run server.")
		return err
	}

	return nil
}
