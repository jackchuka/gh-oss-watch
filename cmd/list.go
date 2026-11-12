package cmd

import (
	"strings"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var lang string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleList(services.NewConfigService(), newFormatter(), lang, tagFilter)
	},
}

func init() {
	listCmd.Flags().StringVarP(&lang, "lang", "l", "", "Filter by language (case-insensitive)")
	addTagFilterFlag(listCmd)
	rootCmd.AddCommand(listCmd)
}

func handleList(configService services.ConfigService, formatter services.Formatter, langFilter string, tags []string) error {
	config, err := configService.Load()
	if err != nil {
		return err
	}

	repos, err := config.FilterByTags(tags)
	if err != nil {
		return err
	}
	if langFilter != "" {
		filtered := make([]services.RepoConfig, 0, len(repos))
		for _, r := range repos {
			if strings.EqualFold(r.Language, langFilter) {
				filtered = append(filtered, r)
			}
		}
		repos = filtered
	}

	return formatter.RenderList(services.ListResult{
		Repos: repos,
		Total: len(repos),
	})
}
