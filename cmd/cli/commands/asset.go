package commands

import (
	"github.com/spf13/cobra"

	"github.com/UnivocalX/aether/internal/actions"
)

// AssetsCmd represents the assets command
var AssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage and interact with data assets.",
	Run:   func(cmd *cobra.Command, args []string) {},
}

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
	client, err := actions.NewClient(
		actions.Interactive(),
	)

	if err != nil {
		return err
	}

	return client.LoadAssets(args[0])
}

func init() {
	AssetsCmd.AddCommand(loadCmd)
}
