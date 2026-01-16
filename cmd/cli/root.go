package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/UnivocalX/aether/cmd/cli/assets"
	"github.com/UnivocalX/aether/internal/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envPrefix     = "AETHER"
	configName    = "config"
	configType    = "yaml"
	configDirName = ".aether"
)

var (
	cfgFile  string
	logLevel string
	endpoint string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "aether",
	Short:             "Aether data platform CLI",
	Long:              "A CLI tool for managing the Aether data platform",
	PersistentPreRunE: initConfig,
}

// logging configuration
var Log = logging.NewLog()

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	Log.Apply()

	// Add subcommands
	rootCmd.AddCommand(assets.AssetsCmd)

	// Define persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aether/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "localhost:8080", "aether API endpoint")

	// Set bash completion for log level
	if err := rootCmd.PersistentFlags().SetAnnotation("level", cobra.BashCompOneRequiredFlag, []string{"debug", "info", "warn", "error"}); err != nil {
		slog.Warn("failed to set bash completion annotation", "error", err)
	}
}

// initConfig is called before command execution to set up configuration
func initConfig(cmd *cobra.Command, args []string) error {
	if err := loadConfig(cmd); err != nil {
		return fmt.Errorf("failed to setup configuration: %w", err)
	}

	// Switch logging to cli
	Log.SetLevel(viper.GetString("level"))
	Log.SetMode(logging.CLIMode)
	Log.Apply()

	slog.Debug("configuration loaded",
		"configFile", viper.ConfigFileUsed(),
		"level", viper.GetString("level"),
		"endpoint", viper.GetString("endpoint"),
	)

	return nil
}

// loadConfig configures viper to read from config files, environment variables, and flags
func loadConfig(cmd *cobra.Command) error {
	// Configure environment variable handling
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// Configure config file
	if err := readConfigFile(); err != nil {
		return err
	}

	// Bind command flags to viper
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind flags: %w", err)
	}

	return nil
}

// readConfigFile handles config file discovery and reading
func readConfigFile() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in default locations
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Add config search paths
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/" + configDirName)
		viper.SetConfigName(configName)
		viper.SetConfigType(configType)
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}
