package assets

import (
	"github.com/spf13/cobra"

	"github.com/UnivocalX/aether/internal/actions"
)

// buildCmd represents the load command
var manifestCmd = &cobra.Command{
	Use:           "manifest <path> ",
	Short:         "manifest an assets manifest file",
	Long:          "manifest an assets manifest file",
	Example:       "aether assets manifest data/ assets.yaml",
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveFilterDirs
	},
	RunE: runBuildManifest,
}

func runBuildManifest(cmd *cobra.Command, args []string) error {
	return actions.BuildManifest(args[0], args[1])
}

func init() {
	AssetsCmd.AddCommand(manifestCmd)
}
