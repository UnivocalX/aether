package assets

import (
	"github.com/spf13/cobra"
)

// AssetsCmd represents the assets command
var AssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage and interact with data assets.",
	Run:   func(cmd *cobra.Command, args []string) {},
}
