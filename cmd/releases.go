package cmd

import (
	"fmt"
	"strings"
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

	c.output.Println("\n\U0001F4E6 Release Status")
	c.output.Println("================")

	// Calculate column widths
	repoWidth := len("Repository")
	releaseWidth := len("Last Release")
	for _, e := range processor.entries {
		if len(e.repo) > repoWidth {
			repoWidth = len(e.repo)
		}
		if len(e.latestRelease) > releaseWidth {
			releaseWidth = len(e.latestRelease)
		}
	}

	// Print header
	c.output.Printf("\n %-*s  %-*s  %-10s  %-13s  %s\n",
		repoWidth, "Repository",
		releaseWidth, "Last Release",
		"Age",
		"Unreleased",
		"Status",
	)
	c.output.Println(strings.Repeat("\u2500", repoWidth+releaseWidth+50))

	needsRelease := 0
	upToDate := 0
	noReleases := 0

	for _, e := range processor.entries {
		var age, unreleased, status string

		switch {
		case e.latestRelease == "":
			age = "\u2014"
			unreleased = "\u2014"
			status = "no releases"
			noReleases++
		case e.unreleasedCount > 0:
			age = humanizeAge(e.releaseDate)
			unreleased = fmt.Sprintf("%d commits", e.unreleasedCount)
			status = "needs release"
			needsRelease++
		default:
			age = humanizeAge(e.releaseDate)
			unreleased = "0 commits"
			status = "up to date"
			upToDate++
		}

		if onlyUnreleased && status != "needs release" {
			continue
		}

		release := e.latestRelease
		if release == "" {
			release = "\u2014"
		}

		c.output.Printf(" %-*s  %-*s  %-10s  %-13s  %s\n",
			repoWidth, e.repo,
			releaseWidth, release,
			age,
			unreleased,
			status,
		)
	}

	// Summary line
	parts := []string{}
	if needsRelease > 0 {
		parts = append(parts, fmt.Sprintf("%d repos need a release", needsRelease))
	}
	if upToDate > 0 {
		parts = append(parts, fmt.Sprintf("%d up to date", upToDate))
	}
	if noReleases > 0 {
		parts = append(parts, fmt.Sprintf("%d no releases", noReleases))
	}
	if len(parts) > 0 {
		c.output.Printf("\n %s\n", strings.Join(parts, " \u00B7 "))
	}

	return nil
}

func humanizeAge(t time.Time) string {
	if t.IsZero() {
		return "\u2014"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	}
}
