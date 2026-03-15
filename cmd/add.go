package cmd

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <repo> [events...]",
	Short: "Add repo to watch list",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		githubService, err := newGitHubService()
		if err != nil {
			return err
		}
		return handleConfigAdd(services.NewConfigService(), githubService, args[0], args[1:])
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func handleConfigAdd(configService services.ConfigService, githubService services.GitHubService, repo string, eventArgs []string) error {
	config, err := configService.Load()
	if err != nil {
		return err
	}

	events := []string{"stars", "issues", "pull_requests", "forks"}
	if len(eventArgs) > 0 {
		events = eventArgs
	}

	owner, repoName, err := services.ParseRepoString(repo)
	if err != nil {
		return err
	}

	if err := githubService.RepoExists(owner, repoName); err != nil {
		return fmt.Errorf("repository does not exist or is inaccessible: %w", err)
	}

	normalized := owner + "/" + repoName

	if err := config.AddRepo(normalized, events); err != nil {
		return err
	}

	err = configService.Save(config)
	if err != nil {
		return err
	}

	fmt.Printf("Added %s to watch list with events: %s\n", normalized, strings.Join(events, ", "))
	return nil
}
