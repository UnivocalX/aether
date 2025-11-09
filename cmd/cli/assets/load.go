package assets

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/actions"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "load files as assets to aether data platform.",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return actions.NewLoadAssets(viper.GetString("endpoint")).Execute(args[0])
	},
}

func init() {
	AssetsCmd.AddCommand(loadCmd)
}
