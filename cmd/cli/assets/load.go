package assets

import (
	"fmt"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "load files as assets to aether data platform.",
	
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("load called")
	},
}

func init() {
	AssetsCmd.AddCommand(loadCmd)
}
