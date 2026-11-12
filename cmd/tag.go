package cmd

import (
	"fmt"
	"strings"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag <tag> <repo...>",
	Short: "Add a tag to repos",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleTag(services.NewConfigService(), args[0], args[1:], true)
	},
}

var untagCmd = &cobra.Command{
	Use:   "untag <tag> <repo...>",
	Short: "Remove a tag from repos",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleTag(services.NewConfigService(), args[0], args[1:], false)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(untagCmd)
}

func handleTag(configService services.ConfigService, tagArg string, repos []string, add bool) error {
	tags, err := services.NormalizeTags([]string{tagArg})
	if err != nil {
		return err
	}

	config, err := configService.Load()
	if err != nil {
		return err
	}

	// Validate every repo before changing anything so a typo leaves the config untouched.
	var missing []string
	for _, repo := range repos {
		if config.GetRepo(repo) == nil {
			missing = append(missing, repo)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("not in watch list: %s", strings.Join(missing, ", "))
	}

	for _, repo := range repos {
		if add {
			err = config.AddTags(repo, tags)
		} else {
			err = config.RemoveTags(repo, tags)
		}
		if err != nil {
			return err
		}
	}

	if err := configService.Save(config); err != nil {
		return err
	}

	if add {
		fmt.Printf("Tagged %d repo(s) with %q\n", len(repos), tags[0])
	} else {
		fmt.Printf("Removed tag %q from %d repo(s)\n", tags[0], len(repos))
	}
	return nil
}
