package cmd

import (
	"github.com/UnivocalX/aether/internal/logger"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show aether version.",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Aether", "Version", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
