package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/api"
	"github.com/UnivocalX/aether/internal/logging"
	"github.com/UnivocalX/aether/pkg/registry"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:           "serve",
	Short:         "Starts the aether api server",
	RunE:          startServer,
	SilenceUsage: true,
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
	serveCmd.Flags().Bool("production", false, "Run in production mode (enables JSON logging)")
}

func LoadAPIConfig() api.Config {
	registryCFG := registry.NewConfig()

	// Storage
	registryCFG.Storage.S3Endpoint = viper.GetString("s3endpoint")
	registryCFG.Storage.Bucket = viper.GetString("bucket")
	registryCFG.Storage.Prefix = viper.GetString("Prefix")

	// Datastore
	registryCFG.Datastore.Endpoint = registry.Endpoint(viper.GetString("db-endpoint"))
	registryCFG.Datastore.User = viper.GetString("db-user")
	registryCFG.Datastore.Password = registry.Secret(viper.GetString("db-password"))
	registryCFG.Datastore.Name = viper.GetString("db-name")
	registryCFG.Datastore.SSL = viper.GetBool("ssl")

	return api.Config{
		Registry: registryCFG,
		Logging: &logging.Config{
			Prod:     viper.GetBool("production"),
			AppName:  "aether",
			LogLevel: logging.LevelFromString("debug"),
		},
		Port: viper.GetString("port"),
	}
}

func startServer(cmd *cobra.Command, args []string) error {
	cfg := LoadAPIConfig()
	slog.Info("Starting server...")

	// Create API
	router, err := api.New(&cfg)
	if err != nil {
		return fmt.Errorf("failed to start server")
	}

	// Run Server
	slog.Info("serving:", "port", cfg.Port, "production", cfg.Logging.Prod)
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("failed to run server.")
		return err
	}

	return nil
}
