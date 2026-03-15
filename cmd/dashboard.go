package cmd

import (
	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

type dashboardProcessor struct {
	entries []services.DashboardEntry
	totals  services.DashboardTotals
}

func (d *dashboardProcessor) ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error {
	d.entries = append(d.entries, services.DashboardEntry{
		Repo:            repoConfig.Repo,
		Stars:           stats.Stars,
		Issues:          stats.Issues,
		PullRequests:    stats.PullRequests,
		Forks:           stats.Forks,
		UpdatedAt:       stats.UpdatedAt,
		LatestRelease:   stats.LatestRelease,
		UnreleasedCount: stats.UnreleasedCount,
		Watching:        repoConfig.Events,
	})

	d.totals.Stars += stats.Stars
	d.totals.Issues += stats.Issues
	d.totals.PRs += stats.PullRequests
	d.totals.Forks += stats.Forks
	if stats.LatestRelease != "" && stats.UnreleasedCount > 0 {
		d.totals.NeedRelease++
	}

	return nil
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show summary across all repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		configService, _, githubService, formatter, err := getServices()
		if err != nil {
			return err
		}
		return handleDashboard(configService, githubService, formatter)
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func handleDashboard(configService services.ConfigService, githubService services.GitHubService, formatter services.Formatter) error {
	config, err := validateConfig(configService)
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	processor := &dashboardProcessor{}

	err = processReposWithBatch(githubService, config, processor)
	if err != nil {
		return err
	}

	return formatter.RenderDashboard(services.DashboardResult{
		Repos:  processor.entries,
		Totals: processor.totals,
	})
}
