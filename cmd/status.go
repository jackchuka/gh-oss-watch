package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

type statusProcessor struct {
	cache   *services.CacheData
	entries []services.StatusEntry
}

func (s *statusProcessor) ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error {
	previousState, exists := s.cache.Repos[repoConfig.Repo]
	if !exists {
		previousState = services.RepoState{}
	}

	summary := services.CalculateEventSummary(repoConfig.Repo, stats, previousState)

	hasVisibleChanges := false
	for _, event := range repoConfig.Events {
		switch event {
		case "stars":
			hasVisibleChanges = hasVisibleChanges || summary.NewStars > 0
		case "issues":
			hasVisibleChanges = hasVisibleChanges || summary.NewIssues > 0
		case "pull_requests":
			hasVisibleChanges = hasVisibleChanges || summary.NewPRs > 0
		case "forks":
			hasVisibleChanges = hasVisibleChanges || summary.NewForks > 0
		case "releases":
			hasVisibleChanges = hasVisibleChanges || summary.NewRelease || summary.UnreleasedCount > 0
		}
	}

	if hasVisibleChanges {
		s.entries = append(s.entries, services.StatusEntry{
			EventSummary: summary,
			Events:       repoConfig.Events,
			TotalStars:   stats.Stars,
			TotalIssues:  stats.Issues,
			TotalPRs:     stats.PullRequests,
			TotalForks:   stats.Forks,
		})
	}

	s.cache.Repos[repoConfig.Repo] = services.RepoState{
		LastStarCount:  stats.Stars,
		LastIssueCount: stats.Issues,
		LastPRCount:    stats.PullRequests,
		LastForkCount:  stats.Forks,
		LastUpdated:    stats.UpdatedAt,
		LastReleaseTag: stats.LatestRelease,
	}

	return nil
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show new activity since last check",
	RunE: func(cmd *cobra.Command, args []string) error {
		configService, cacheService, githubService, formatter, err := getServices()
		if err != nil {
			return err
		}
		return handleStatus(configService, cacheService, githubService, formatter)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func handleStatus(configService services.ConfigService, cacheService services.CacheService, githubService services.GitHubService, formatter services.Formatter) error {
	config, err := validateConfig(configService)
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	cache, err := cacheService.Load()
	if err != nil {
		return err
	}

	processor := &statusProcessor{
		cache: cache,
	}

	err = processReposWithBatch(githubService, config, processor)
	if err != nil {
		return err
	}

	if err := formatter.RenderStatus(processor.entries); err != nil {
		return err
	}

	cache.LastCheck = time.Now()
	err = cacheService.Save(cache)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error saving cache: %v\n", err)
	}

	return nil
}
