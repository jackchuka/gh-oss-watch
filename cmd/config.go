package cmd

import (
	"fmt"
	"github.com/jackchuka/gh-oss-watch/services"
	"strings"
)

func (c *CLI) handleConfigAdd(repo string, eventArgs []string) error {
	config, err := c.configService.Load()
	if err != nil {
		return err
	}

	events := []string{"stars", "issues", "pull_requests", "forks"}
	if len(eventArgs) > 0 {
		events = eventArgs
	}

	owner, repo, err := services.ParseRepoString(repo)
	if err != nil {
		return err
	}

	exists, err := c.githubService.RepoExists(owner, repo)

	if !exists {

		return fmt.Errorf("github repository not found")
	}

	if err := config.AddRepo(repo, events); err != nil {
		return err
	}

	err = c.configService.Save(config)
	if err != nil {
		return err
	}

	c.output.Printf("Added %s to watch list with events: %s\n", repo, strings.Join(events, ", "))
	return nil
}

func (c *CLI) handleConfigSet(repo string, eventArgs []string) error {
	if len(eventArgs) == 0 {
		return fmt.Errorf("no events specified")
	}

	config, err := c.configService.Load()
	if err != nil {
		return err
	}

	repoConfig := config.GetRepo(repo)
	if repoConfig == nil {
		return fmt.Errorf("repository %s not found in config. Use 'gh oss-watch add' first", repo)
	}

	if err := config.AddRepo(repo, eventArgs); err != nil {
		return err
	}

	err = c.configService.Save(config)
	if err != nil {
		return err
	}

	c.output.Printf("Updated %s events to: %s\n", repo, strings.Join(eventArgs, ", "))
	return nil
}

func (c *CLI) handleConfigRemove(repo string) error {
	config, err := c.configService.Load()
	if err != nil {
		return err
	}

	if err := config.RemoveRepo(repo); err != nil {
		return err
	}

	err = c.configService.Save(config)
	if err != nil {
		return err
	}

	c.output.Printf("Removed %s from watch list\n", repo)
	return nil
}
