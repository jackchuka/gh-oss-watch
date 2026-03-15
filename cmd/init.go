package cmd

import (
	"fmt"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleInit(services.NewConfigService())
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func handleInit(configService services.ConfigService) error {
	config, err := configService.Load()
	if err != nil {
		return err
	}

	configPath, err := configService.GetConfigPath()
	if err != nil {
		return err
	}

	err = configService.Save(config)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized config file at %s\n", configPath)
	fmt.Println("Use 'gh oss-watch add <repo>' to start watching repositories")
	return nil
}
