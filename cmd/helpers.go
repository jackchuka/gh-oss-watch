package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
)

var (
	format        string
	maxConcurrent int
	timeout       int
)

func getServices() (services.ConfigService, services.CacheService, services.BatchGitHubService, services.Formatter, error) {
	configService := services.NewConfigService()
	cacheService := services.NewCacheService()
	githubService, err := services.NewConcurrentGitHubService()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	githubService.SetMaxConcurrent(maxConcurrent)
	githubService.SetTimeout(time.Duration(timeout) * time.Second)
	formatter := services.NewFormatter(format)
	return configService, cacheService, githubService, formatter, nil
}
