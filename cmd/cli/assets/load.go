package assets

import (
	"github.com/spf13/cobra"

	"github.com/UnivocalX/aether/internal/client"
)

// buildCmd represents the load command
var loadCmd = &cobra.Command{
	Use:           "load <path> ",
	Short:         "load assets",
	Long:          "load assets to aether platform",
	Example:       "aether assets load data/ ",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveFilterDirs
	},
	RunE: runLoadAssets,
}

func runLoadAssets(cmd *cobra.Command, args []string) error {
	return client.LoadAssets(args[0], true)
}

func init() {
	AssetsCmd.AddCommand(loadCmd)
}
