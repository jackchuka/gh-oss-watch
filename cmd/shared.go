package cmd

import (
	"fmt"
	"os"

	"github.com/jackchuka/gh-oss-watch/services"
)

func validateConfig(configService services.ConfigService) (*services.Config, error) {
	config, err := configService.Load()
	if err != nil {
		return nil, err
	}

	if len(config.Repos) == 0 {
		fmt.Println("No repositories configured. Use 'gh oss-watch add <repo>' to add some.")
		return config, nil
	}

	return config, nil
}

type RepoStatsProcessor interface {
	ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error
}

func processReposWithBatch(
	githubService services.GitHubService,
	config *services.Config,
	processor RepoStatsProcessor,
) error {
	batchService, canBatch := githubService.(services.BatchGitHubService)
	if !canBatch || len(config.Repos) <= 1 {
		return processReposSequentially(githubService, config, processor)
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

func backfillLanguages(configService services.ConfigService, config *services.Config, allStats []*services.RepoStats) {
	needsSave := false
	for i, repoConfig := range config.Repos {
		if repoConfig.Language != "" {
			continue
		}
		if i < len(allStats) && allStats[i] != nil && allStats[i].Language != "" {
			config.Repos[i].Language = allStats[i].Language
			needsSave = true
		}
	}
	if needsSave {
		if err := configService.Save(config); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to backfill language info: %v\n", err)
		}
	}
}

func processReposSequentially(
	githubService services.GitHubService,
	config *services.Config,
	processor RepoStatsProcessor,
) error {
	for i, repoConfig := range config.Repos {
		owner, repo, err := services.ParseRepoString(repoConfig.Repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing repo %s: %v\n", repoConfig.Repo, err)
			continue
		}

		stats, err := githubService.GetRepoStats(owner, repo)
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
