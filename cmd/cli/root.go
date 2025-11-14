package main

import (
	"log/slog"
	"os"

	"github.com/UnivocalX/aether/internal/logging"
	"github.com/UnivocalX/aether/internal/settings"
	"github.com/UnivocalX/aether/cmd/cli/assets"

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
	rootCmd.AddCommand(assets.AssetsCmd)

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.aether/config.yaml)")
	rootCmd.PersistentFlags().String("level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("endpoint", "localhost:8080", "aether api endpoint.")
	rootCmd.PersistentFlags().SetAnnotation("level", cobra.BashCompOneRequiredFlag, []string{"debug", "info", "warn", "error"})
}


func Setup(cmd *cobra.Command, args []string) error {
	// Initialize settings first
	if err := settings.Setup(cmd); err != nil {
		return err
	}

	// Setup logging
	logger := logging.New(nil, logging.LevelFromString(viper.GetString("level")))
	slog.SetDefault(logger)

	slog.Debug("Configuration loaded",
		"ConfigFile", viper.ConfigFileUsed(),
		"Level", viper.GetString("level"),
	)
	return nil
}
