package cmd

import (
	"fmt"
	"os"

	"github.com/jackchuka/gh-oss-watch/services"
)

func (c *CLI) validateConfig() (*services.Config, error) {
	config, err := c.configService.Load()
	if err != nil {
		return nil, err
	}

	if len(config.Repos) == 0 {
		c.output.Println("No repositories configured. Use 'gh oss-watch add <repo>' to add some.")
		return config, nil
	}

	return config, nil
}

type RepoStatsProcessor interface {
	ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error
}

func (c *CLI) processReposWithBatch(
	config *services.Config,
	processor RepoStatsProcessor,
) error {
	batchService, canBatch := c.githubService.(services.BatchGitHubService)
	if !canBatch || len(config.Repos) <= 1 {
		return c.processReposSequentially(config, processor)
	}

	repos := make([]string, len(config.Repos))
	for i, repoConfig := range config.Repos {
		repos[i] = repoConfig.Repo
	}

	allStats, allErrors := batchService.GetRepoStatsBatch(repos)

	for i, repoConfig := range config.Repos {
		if allErrors[i] != nil {
			fmt.Fprintf(os.Stderr, "Error fetching stats for %s: %v\n", repoConfig.Repo, allErrors[i])
			continue
		}

		stats := allStats[i]
		if stats == nil {
			continue
		}

		if err := processor.ProcessRepo(repoConfig, stats, i); err != nil {
			return err
		}
	}

	return nil
}

func (c *CLI) processReposSequentially(
	config *services.Config,
	processor RepoStatsProcessor,
) error {
	for i, repoConfig := range config.Repos {
		owner, repo, err := services.ParseRepoString(repoConfig.Repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing repo %s: %v\n", repoConfig.Repo, err)
			continue
		}

		stats, err := c.githubService.GetRepoStats(owner, repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching stats for %s: %v\n", repoConfig.Repo, err)
			continue
		}

		if err := processor.ProcessRepo(repoConfig, stats, i); err != nil {
			return err
		}
	}

	return nil
}
