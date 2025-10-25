package cmd

import (
	"log/slog"
	"os"

	"github.com/UnivocalX/aether/internal/logger"
	"github.com/UnivocalX/aether/internal/settings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "aether",
	Short:             "Aether data platform CLI.",
	PersistentPreRunE: Setup,
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
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.aether.yaml)")
	rootCmd.PersistentFlags().Bool("production", false, "Run in production mode (enables JSON logging)")
	rootCmd.PersistentFlags().String("level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("endpoint", "", "aether api endpoint.")
	rootCmd.PersistentFlags().SetAnnotation("level", cobra.BashCompOneRequiredFlag, []string{"debug", "info", "warn", "error"})
}

func Setup(cmd *cobra.Command, args []string) error {
	// Initialize settings first
	if err := settings.Setup(cmd); err != nil {
		return err
	}

	// Setup logging
	if err := logger.Setup(&logger.Options{
		AddSource:  true,
		Production: viper.GetBool("production"),
		Level:      viper.GetString("level"),
	}); err != nil {
		logger.SetupDefault()
		slog.Error("Failed to set up logger", "error", err)
	}

	slog.Debug("Configuration loaded",
		"ConfigFile", viper.ConfigFileUsed(),
		"Level", viper.GetString("level"),
		"Production", viper.GetBool("production"),
	)

	return nil
}
