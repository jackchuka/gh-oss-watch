package cmd

import (
	"time"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var onlyUnreleased bool

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

var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Show release status across all repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		configService, _, githubService, formatter, err := getServices()
		if err != nil {
			return err
		}
		return handleReleases(configService, githubService, formatter, onlyUnreleased)
	},
}

func init() {
	releasesCmd.Flags().BoolVarP(&onlyUnreleased, "only-unreleased", "u", false, "Show only repos that need a release")
	rootCmd.AddCommand(releasesCmd)
}

func handleReleases(configService services.ConfigService, githubService services.GitHubService, formatter services.Formatter, onlyUnreleased bool) error {
	config, err := validateConfig(configService)
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	processor := &releasesProcessor{}

	err = processReposWithBatch(githubService, config, processor)
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

	return formatter.RenderReleases(releases)
}
