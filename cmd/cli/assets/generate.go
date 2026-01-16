package assets

import (
	"github.com/spf13/cobra"
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
	RunE: GenerateManifest,
}

func GenerateManifest(cmd *cobra.Command, args []string) error {
	return nil
}

func init() {
	AssetsCmd.AddCommand(genCmd)
}
