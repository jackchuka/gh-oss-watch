package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-oss-watch",
	Short: "GitHub CLI extension for OSS maintainers",
	Long:  "Monitor stars, issues, PRs, forks, and releases across your open-source repositories.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "plain", "Output format: plain, json")
	rootCmd.PersistentFlags().IntVar(&maxConcurrent, "max-concurrent", 10, "Max concurrent API requests")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")
}
