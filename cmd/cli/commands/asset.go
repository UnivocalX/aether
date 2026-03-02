package commands

import (
	"time"

	"github.com/UnivocalX/aether/pkg/client"
	"github.com/spf13/cobra"
)

const (
	DefaultTimeoutSeconds = 3600
)

// AssetsCmd represents the assets command
var AssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage and interact with data assets.",
}

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:           "load <path>",
	Short:         "Load assets",
	Long:          "Load assets to the Aether platform",
	Example:       "aether assets load data/",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt | cobra.ShellCompDirectiveFilterDirs
	},
	RunE: runLoadAssets,
}

func init() {
	AssetsCmd.AddCommand(loadCmd)
	AssetsCmd.PersistentFlags().Bool("ci", false, "Disable user interaction and progress bars")
	AssetsCmd.PersistentFlags().Int("timeout", DefaultTimeoutSeconds, "Command timeout in seconds")
}

func runLoadAssets(cmd *cobra.Command, args []string) error {
	ci, _ := cmd.Flags().GetBool("ci")
	timeout, _ := cmd.Flags().GetInt("timeout")
	endpoint, _ := cmd.Flags().GetString("endpoint")

	aether, err := client.New(
		client.WithMode(!ci),
		client.WithTimeout(time.Duration(timeout)*time.Second),
		client.WithEndpoint(endpoint),
	)

	if err != nil {
		return err
	}

	return aether.LoadAssets(args[0])
}
