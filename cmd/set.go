package cmd

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <repo> <events...>",
	Short: "Configure events for repo",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configService, _, _, _, err := getServices()
		if err != nil {
			return err
		}
		return handleConfigSet(configService, args[0], args[1:])
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}

func handleConfigSet(configService services.ConfigService, repo string, eventArgs []string) error {
	if len(eventArgs) == 0 {
		return fmt.Errorf("no events specified")
	}

	config, err := configService.Load()
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

	err = configService.Save(config)
	if err != nil {
		return err
	}

	fmt.Printf("Updated %s events to: %s\n", repo, strings.Join(eventArgs, ", "))
	return nil
}
