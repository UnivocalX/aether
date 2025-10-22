package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the aether api server",
	RunE: func(cmd *cobra.Command, args []string) error {
			port := viper.GetInt("port")
			slog.Info("Starting aether server.", "Port", port,)
		return nil
	},
}


func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to run the server on")
}
