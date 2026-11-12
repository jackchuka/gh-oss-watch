package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var (
	format        string
	maxConcurrent int
	timeout       int
	noCache       bool
	tagFilter     []string
)

func addTagFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&tagFilter, "tag", "T", nil, "Only include repos with any of these tags (repeatable or comma-separated)")
}

func newGitHubService() (services.BatchGitHubService, error) {
	githubService, err := services.NewConcurrentGitHubService(!noCache)
	if err != nil {
		return nil, err
	}
	githubService.SetMaxConcurrent(maxConcurrent)
	githubService.SetTimeout(time.Duration(timeout) * time.Second)
	return githubService, nil
}

func newFormatter() services.Formatter {
	return services.NewFormatter(format)
}
