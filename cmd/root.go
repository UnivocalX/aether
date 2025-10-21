package cmd

import (
	"os"

	"github.com/UnivocalX/aether/internal/settings"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aether",
	Short: "Aether data platform CLI.",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return settings.Init(cmd)
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
	rootCmd.PersistentFlags().SetAnnotation("level", cobra.BashCompOneRequiredFlag, []string{"debug", "info", "warn", "error"})
}
