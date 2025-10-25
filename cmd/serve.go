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
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to run the server on")
	serveCmd.Flags().String("s3endpoint", "", "Port to run the server on")
	serveCmd.Flags().String("bucket", "", "Port to run the server on")
	serveCmd.Flags().String("prefix", "aether", "Port to run the server on")
}

func startServer(cmd *cobra.Command, args []string) error {
	opt := api.Options{
		Registry: &registry.Options{
			S3Endpoint: viper.GetString("s3endpoint"),
			Bucket:     viper.GetString("bucket"),
			Prefix:     viper.GetString("prefix"),
		},
		Production: viper.GetBool("production"),
	}

	router, err := api.New(&opt)
	if err != nil {
		return err
	}

	slog.Info("Starting API server",
		"port", viper.GetInt("port"),
		"s3endpoint", opt.Registry.S3Endpoint,
		"bucket", opt.Registry.Bucket,
		"prefix", opt.Registry.Prefix,
		"production", opt.Production,
	)

	err = router.Run(":" + viper.GetString("port"))
	if err != nil {
		return err
	}

	return nil
}
