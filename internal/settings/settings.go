package settings

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/UnivocalX/aether/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Options holds all application configuration with embedded logger options
type Options struct {
	Logging    logger.Options
	ConfigPath string
}

// Init initializes the configuration system and returns the config
func Init(cmd *cobra.Command) error {
	// 1. Set up Viper to use environment variables.
	viper.SetEnvPrefix("AETHER")

	// Allow for nested keys in environment variables (e.g. `AETHER_DATABASE_HOST`)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	// 2. Handle the configuration file.
	cfgFile, _ := cmd.Flags().GetString("config")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for a config file in default locations.
		home, err := os.UserHomeDir()
		// Only panic if we can't get the home directory.
		if err != nil {
			return err
		}

		// Search for a config file with the name "config" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.aether")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// 3. Read the configuration file.
	// If a config file is found, read it in. We use a robust error check
	// to ignore "file not found" errors, but panic on any other error.
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if the config file doesn't exist.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	// 4. Bind Cobra flags to Viper.
	// This is the magic that makes the flag values available through Viper.
	// It binds the full flag set of the command passed in.
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	// 5. Validate log level before initializing logger
	logLevel := viper.GetString("level")
	if err := validateLogLevel(logLevel); err != nil {
		logLevel = "info" // Fallback to default
	}

	// 6. Create config options with embedded logger options
	opts := &Options{
		Logging: logger.Options{
			AddSource:  true, // You can make this configurable too
			Production: viper.GetBool("production"),
			Level:      logLevel,
		},
		ConfigPath: viper.ConfigFileUsed(),
	}

	// 7. Initialize logger once with the config
	logger.Init(opts.Logging)

	// 8. Log the configuration initialization
	logger.Debug("Configuration initialized.",
		"Config Path", opts.ConfigPath,
		"Production", opts.Logging.Production,
		"Log Level", opts.Logging.Level,
	)

	return nil
}

// validateLogLevel validates that the log level is one of the allowed values
func validateLogLevel(level string) error {
	allowedLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true,
		"error":   true,
	}

	levelLower := strings.ToLower(level)
	if !allowedLevels[levelLower] {
		return fmt.Errorf("invalid log level %q, must be one of: debug, info, warn, error", level)
	}

	return nil
}
