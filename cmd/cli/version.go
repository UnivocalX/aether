package main

import (
	"runtime"

	"github.com/spf13/cobra"
	"log/slog"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

type BuildInfo struct {
	Version    string
	Commit     string
	Date       string
	TargetOS   string
	TargetArch string
	GoVersion  string
}

func GetBuildInfo() BuildInfo {
	build := BuildInfo{
		Version:    version,
		Commit:     commit,
		Date:       date,
		TargetOS:   runtime.GOOS,
		TargetArch: runtime.GOARCH,
		GoVersion:  runtime.Version(),
	}

	slog.Info("Aether", 
		slog.String("Version", build.Version),
		slog.String("Commit", build.Commit), 
		slog.String("Build Date", build.Date), 
		slog.String("Target OS", build.TargetOS), 
		slog.String("Target Architecture", build.TargetArch), 
		slog.String("Go", build.GoVersion),
	)
	return build
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show aether version.",
	Run: func(cmd *cobra.Command, args []string) {
		GetBuildInfo()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
