package cmd

import (
	"os"

	"github.com/UnivocalX/aether/internal/settings"
	"github.com/UnivocalX/aether/pkg/registry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	logLevel string
	registryClient   *registry.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aether",
	Short: "Aether data platform CLI.",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize settings first
		err := settings.Init(cmd)
		if err != nil {
			return err
		}

		// Create config from viper
		cfg := registry.Config{
			S3Endpoint: viper.GetString("s3endpoint"),
			Bucket:     viper.GetString("bucket"),
			Prefix:     viper.GetString("prefix"),
		}

		// Initialize client and assign to package variable (no shadowing)
		registryClient, err = registry.New(cfg)
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aether.yaml)")

	rootCmd.PersistentFlags().Bool("production", false, "Run in production mode (enables JSON logging)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("s3endpoint", "", "Object store endpoint.")
	rootCmd.PersistentFlags().String("bucket", "", "Object store bucket name.")
	rootCmd.PersistentFlags().String("prefix", "aether", "Object store path prefix.")
	rootCmd.PersistentFlags().SetAnnotation("level", cobra.BashCompOneRequiredFlag, []string{"debug", "info", "warn", "error"})
}

// GetRegistryClient returns the shared client instance for subcommands
func GetRegistryClient() *registry.Client {
	return registryClient
}
