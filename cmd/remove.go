package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <repo>",
	Short: "Remove repo from watch list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleConfigRemove(services.NewConfigService(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func handleConfigRemove(configService services.ConfigService, repo string) error {
	config, err := configService.Load()
	if err != nil {
		return err
	}

	if err := config.RemoveRepo(repo); err != nil {
		return err
	}

	err = configService.Save(config)
	if err != nil {
		return err
	}

	fmt.Printf("Removed %s from watch list\n", repo)
	return nil
}
