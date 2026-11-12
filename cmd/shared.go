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

func repoNames(repos []services.RepoConfig) []string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Repo
	}
	return names
}

func processReposWithBatch(
	githubService services.GitHubService,
	repos []services.RepoConfig,
	processor RepoStatsProcessor,
) error {
	batchService, canBatch := githubService.(services.BatchGitHubService)
	if !canBatch || len(repos) <= 1 {
		return processReposSequentially(githubService, repos, processor)
	}

	allStats, allErrors := batchService.GetRepoStatsBatch(repoNames(repos))

	for i, repoConfig := range repos {
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

// backfillLanguages saves languages learned from stats into the full config.
// repos is the (possibly tag-filtered) list that allStats is indexed by, so the
// config is updated by repo name rather than by position.
func backfillLanguages(configService services.ConfigService, config *services.Config, repos []services.RepoConfig, allStats []*services.RepoStats) {
	learned := make(map[string]string)
	for i, repoConfig := range repos {
		if repoConfig.Language == "" && i < len(allStats) && allStats[i] != nil && allStats[i].Language != "" {
			learned[repoConfig.Repo] = allStats[i].Language
		}
	}

	needsSave := false
	for i, repoConfig := range config.Repos {
		if lang, ok := learned[repoConfig.Repo]; ok && repoConfig.Language == "" {
			config.Repos[i].Language = lang
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
	repos []services.RepoConfig,
	processor RepoStatsProcessor,
) error {
	for i, repoConfig := range repos {
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
