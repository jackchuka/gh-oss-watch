package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
)

var (
	format        string
	maxConcurrent int
	timeout       int
	noCache       bool
)

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
