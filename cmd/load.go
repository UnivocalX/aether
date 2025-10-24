package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var loadCmd = &cobra.Command{
	Use:   "load [path-pattern]",
	Short: "Load data to registry.",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		registryClient := GetRegistryClient()
		_, err := registryClient.Load(args[0])
		
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}

