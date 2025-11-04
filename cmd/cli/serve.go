package main

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
	SilenceUsage:  true,
	SilenceErrors: true,
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

func LoadAPIConfig() api.Config {
	registryCFG := registry.NewConfig()

	// Storage - using nested keys
	registryCFG.Storage.S3Endpoint = viper.GetString("storage.s3endpoint")
	registryCFG.Storage.Bucket = viper.GetString("storage.bucket")
	registryCFG.Storage.Prefix = viper.GetString("storage.prefix")

	// Database - using nested keys
	registryCFG.Database.Endpoint = registry.Endpoint(viper.GetString("database.endpoint"))
	registryCFG.Database.User = viper.GetString("database.user")
	registryCFG.Database.Password = registry.Secret(viper.GetString("database.password"))
	registryCFG.Database.Name = viper.GetString("database.name")
	registryCFG.Database.SSL = viper.GetBool("database.ssl")

	return api.Config{
		Registry: registryCFG,
		Logging: &logging.Config{
			Prod:     viper.GetBool("server.production"),
			AppName:  "aether",
			LogLevel: logging.LevelFromString(viper.GetString("level")),
		},
		Port: viper.GetString("server.port"),
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
	if err := router.Run(":" + cfg.Port); err != nil {
		slog.Error("failed to run server.")
		return err
	}

	return nil
}

func bindServeFlags() {
	// Server settings
	viper.BindPFlag("server.port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("server.production", serveCmd.Flags().Lookup("production"))

	// Storage settings
	viper.BindPFlag("storage.s3endpoint", serveCmd.Flags().Lookup("s3endpoint"))
	viper.BindPFlag("storage.bucket", serveCmd.Flags().Lookup("bucket"))
	viper.BindPFlag("storage.prefix", serveCmd.Flags().Lookup("prefix"))

	// Database settings
	viper.BindPFlag("database.endpoint", serveCmd.Flags().Lookup("db-endpoint"))
	viper.BindPFlag("database.user", serveCmd.Flags().Lookup("db-user"))
	viper.BindPFlag("database.password", serveCmd.Flags().Lookup("db-password"))
	viper.BindPFlag("database.name", serveCmd.Flags().Lookup("db-name"))
	viper.BindPFlag("database.ssl", serveCmd.Flags().Lookup("ssl"))
}
