package assets

import (
	"github.com/spf13/cobra"

	"github.com/UnivocalX/aether/internal/actions"
)

// genCmd represents the load command
var genCmd = &cobra.Command{
	Use:           "gen <path> ",
	Short:         "generate assets manifest.",
	Long:          "generate assets manifest.",
	Example:       "aether assets gen data/ manifest.yaml",
	Args:          cobra.MinimumNArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveFilterDirs
	},
	RunE: startGenManifest,
}

func startGenManifest(cmd *cobra.Command, args []string) error {
	return actions.GenerateManifest(args[0], args[1])
}

func init() {
	AssetsCmd.AddCommand(genCmd)
}
