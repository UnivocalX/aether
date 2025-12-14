package assets

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/actions"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load <path> [paths...]",
	Short: "Load files as assets to aether data platform.",
	Long:  "Load files as assets to aether data platform.",
	Example: "aether assets load data/ --tags archive",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveFilterDirs
	},
	RunE: RunLoadAssets,
}

func RunLoadAssets(cmd *cobra.Command, args []string) error {
	load, err := actions.NewLoadAssets(
		actions.WithEndpoint(
			viper.GetString("endpoint"),
		),
	)

	if err != nil {
		return err
	}

	return load.Run(args, viper.GetStringSlice("tags"))
}

func init() {
	AssetsCmd.AddCommand(loadCmd)

	// Add tags flag
	loadCmd.Flags().StringSliceP("tags", "t", []string{},
		"Tags to associate with the assets (can specify multiple)")
}
