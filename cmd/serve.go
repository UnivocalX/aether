package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/UnivocalX/aether/internal/api"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the aether api server",
	RunE:  startServer,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int("port", 8080, "Port to run the server on")
}

func startServer(cmd *cobra.Command, args []string) error {
	router := api.New(viper.GetBool("production"))
	err := router.Run(":" + viper.GetString("port"))

	if err != nil {
		return err
	}

	return nil
}
