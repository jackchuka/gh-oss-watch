package cmd

import (
	"strings"

	"github.com/jackchuka/gh-oss-watch/services"
)

type dashboardProcessor struct {
	output     services.Output
	totalStats *struct {
		Stars       int
		Issues      int
		PRs         int
		Forks       int
		NeedRelease int
	}
}

func (d *dashboardProcessor) ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error {
	d.output.Printf("\n\U0001F4C1 %s\n", repoConfig.Repo)
	d.output.Printf("   \u2B50 Stars: %d\n", stats.Stars)
	d.output.Printf("   \U0001F41B Issues: %d\n", stats.Issues)
	d.output.Printf("   \U0001F500 Pull Requests: %d\n", stats.PullRequests)
	d.output.Printf("   \U0001F374 Forks: %d\n", stats.Forks)
	if stats.LatestRelease != "" {
		d.output.Printf("   \U0001F4E6 Latest Release: %s (%s ago)\n", stats.LatestRelease, humanizeAge(stats.ReleaseDate))
		if stats.UnreleasedCount > 0 {
			d.output.Printf("   \U0001F4E6 Unreleased: %d commits\n", stats.UnreleasedCount)
			d.totalStats.NeedRelease++
		}
	}
	d.output.Printf("   \U0001F4C5 Last Updated: %s\n", stats.UpdatedAt.Format("2006-01-02 15:04"))
	d.output.Printf("   \U0001F4E2 Watching: %s\n", strings.Join(repoConfig.Events, ", "))

	d.totalStats.Stars += stats.Stars
	d.totalStats.Issues += stats.Issues
	d.totalStats.PRs += stats.PullRequests
	d.totalStats.Forks += stats.Forks

	return nil
}

func (c *CLI) handleDashboard() error {
	config, err := c.validateConfig()
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	c.output.Println("\U0001F4CA OSS Watch Dashboard")
	c.output.Println("======================")

	totalStats := struct {
		Stars       int
		Issues      int
		PRs         int
		Forks       int
		NeedRelease int
	}{}

	processor := &dashboardProcessor{
		output:     c.output,
		totalStats: &totalStats,
	}

	err = c.processReposWithBatch(config, processor)
	if err != nil {
		return err
	}

	c.output.Println("\n\U0001F4C8 Total Across All Repos:")
	c.output.Printf("   \u2B50 Total Stars: %d\n", totalStats.Stars)
	c.output.Printf("   \U0001F41B Total Issues: %d\n", totalStats.Issues)
	c.output.Printf("   \U0001F500 Total PRs: %d\n", totalStats.PRs)
	c.output.Printf("   \U0001F374 Total Forks: %d\n", totalStats.Forks)
	if totalStats.NeedRelease > 0 {
		c.output.Printf("   \U0001F4E6 Repos Needing Release: %d\n", totalStats.NeedRelease)
	}

	return nil
}
