package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
)

type releaseEntry struct {
	repo            string
	latestRelease   string
	releaseDate     time.Time
	unreleasedCount int
}

type releasesProcessor struct {
	entries []releaseEntry
}

func (r *releasesProcessor) ProcessRepo(repoConfig services.RepoConfig, stats *services.RepoStats, index int) error {
	r.entries = append(r.entries, releaseEntry{
		repo:            repoConfig.Repo,
		latestRelease:   stats.LatestRelease,
		releaseDate:     stats.ReleaseDate,
		unreleasedCount: stats.UnreleasedCount,
	})
	return nil
}

func (c *CLI) handleReleases(onlyUnreleased bool) error {
	config, err := c.validateConfig()
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	processor := &releasesProcessor{}

	err = c.processReposWithBatch(config, processor)
	if err != nil {
		return err
	}

	var releases []services.ReleaseInfo
	for _, e := range processor.entries {
		var age, status string
		switch {
		case e.latestRelease == "":
			age = "\u2014"
			status = "no releases"
		case e.unreleasedCount > 0:
			age = services.HumanizeAge(e.releaseDate)
			status = "needs release"
		default:
			age = services.HumanizeAge(e.releaseDate)
			status = "up to date"
		}

		ri := services.ReleaseInfo{
			Repo:            e.repo,
			LatestRelease:   e.latestRelease,
			ReleaseDate:     e.releaseDate,
			ReleaseAge:      age,
			UnreleasedCount: e.unreleasedCount,
			Status:          status,
		}

		if onlyUnreleased && ri.Status != "needs release" {
			continue
		}

		releases = append(releases, ri)
	}

	return c.formatter.RenderReleases(releases)
}
