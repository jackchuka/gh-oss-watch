package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
)

type statusProcessor struct {
	output     services.Output
	cache      *services.CacheData
	hasChanges *bool
}

func (s *statusProcessor) ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error {
	previousState, exists := s.cache.Repos[repoConfig.Repo]
	if !exists {
		previousState = services.RepoState{}
	}

	summary := services.CalculateEventSummary(repoConfig.Repo, stats, previousState)

	// Determine if there are visible changes for the configured events
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
		*s.hasChanges = true
		s.output.Printf("\n\U0001F4C8 %s:\n", repoConfig.Repo)

		for _, event := range repoConfig.Events {
			switch event {
			case "stars":
				if summary.NewStars > 0 {
					s.output.Printf("  \u2B50 +%d stars (%d total)\n", summary.NewStars, stats.Stars)
				}
			case "issues":
				if summary.NewIssues > 0 {
					s.output.Printf("  \U0001F41B +%d issues (%d open)\n", summary.NewIssues, stats.Issues)
				}
			case "pull_requests":
				if summary.NewPRs > 0 {
					s.output.Printf("  \U0001F500 +%d pull requests (%d open)\n", summary.NewPRs, stats.PullRequests)
				}
			case "forks":
				if summary.NewForks > 0 {
					s.output.Printf("  \U0001F374 +%d forks (%d total)\n", summary.NewForks, stats.Forks)
				}
			case "releases":
				if summary.NewRelease {
					s.output.Printf("  \U0001F4E6 new release %s\n", summary.ReleaseTag)
				}
				if summary.UnreleasedCount > 0 {
					s.output.Printf("  \U0001F4E6 %d unreleased commits since %s (%s ago)\n",
						summary.UnreleasedCount, summary.ReleaseTag, humanizeAge(stats.ReleaseDate))
				}
			}
		}
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

func (c *CLI) handleStatus() error {
	config, err := c.validateConfig()
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	cache, err := c.cacheService.Load()
	if err != nil {
		return err
	}

	hasChanges := false

	processor := &statusProcessor{
		output:     c.output,
		cache:      cache,
		hasChanges: &hasChanges,
	}

	err = c.processReposWithBatch(config, processor)
	if err != nil {
		return err
	}

	if !hasChanges {
		c.output.Println("No new activity since last check.")
	}

	cache.LastCheck = time.Now()
	err = c.cacheService.Save(cache)
	if err != nil {
		c.output.Printf("Warning: Error saving cache: %v\n", err)
	}

	return nil
}
