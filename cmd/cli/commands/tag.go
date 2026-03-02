package commands

import (
	"github.com/spf13/cobra"
)

// AssetsCmd represents the assets command
var TagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage and interact with data tags.",
}
