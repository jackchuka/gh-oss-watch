package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/jackchuka/gh-oss-watch/services"
	"github.com/spf13/cobra"
)

var (
	securityDetail   bool
	securitySeverity string
	securityRepo     string
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Show open Dependabot security alerts across all repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		githubService, err := newGitHubService()
		if err != nil {
			return err
		}
		return handleSecurity(services.NewConfigService(), githubService, newFormatter(), securityDetail, securitySeverity, securityRepo)
	},
}

func init() {
	securityCmd.Flags().BoolVarP(&securityDetail, "detail", "d", false, "Show every alert, not just per-repo counts")
	securityCmd.Flags().StringVarP(&securitySeverity, "severity", "s", "", "Only show alerts at or above this severity: critical, high, medium, low")
	securityCmd.Flags().StringVarP(&securityRepo, "repo", "r", "", "Limit to a single watched repo (owner/name)")
	rootCmd.AddCommand(securityCmd)
}

func handleSecurity(configService services.ConfigService, githubService services.BatchGitHubService, formatter services.Formatter, detail bool, severityFloor, repoFilter string) error {
	config, err := validateConfig(configService)
	if err != nil {
		return err
	}
	if len(config.Repos) == 0 {
		return nil
	}

	if err := validateSecurityFlags(config, severityFloor, repoFilter); err != nil {
		return err
	}

	repos := selectSecurityRepos(config, repoFilter)

	alertService, ok := githubService.(services.DependabotAlertsBatchService)
	if !ok {
		return fmt.Errorf("github service does not support alert scanning")
	}

	alertsPerRepo, errs := alertService.GetDependabotAlertsBatch(repos)

	for i, repo := range repos {
		if i < len(errs) && errs[i] != nil && !errors.Is(errs[i], services.ErrNoAlertAccess) {
			fmt.Fprintf(os.Stderr, "Error fetching alerts for %s: %v\n", repo, errs[i])
		}
	}

	result := services.BuildSecurityResult(repos, alertsPerRepo, errs, severityFloor)
	return formatter.RenderSecurity(result, detail)
}

func selectSecurityRepos(config *services.Config, repoFilter string) []string {
	if repoFilter != "" {
		return []string{repoFilter}
	}
	repos := make([]string, len(config.Repos))
	for i, rc := range config.Repos {
		repos[i] = rc.Repo
	}
	return repos
}

func validateSecurityFlags(config *services.Config, severityFloor, repoFilter string) error {
	switch severityFloor {
	case "", "critical", "high", "medium", "low":
	default:
		return fmt.Errorf("invalid --severity %q: must be one of critical, high, medium, low", severityFloor)
	}

	if repoFilter != "" {
		found := false
		for _, rc := range config.Repos {
			if rc.Repo == repoFilter {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("repo %q is not in your watch list (use 'gh oss-watch add %s')", repoFilter, repoFilter)
		}
	}

	return nil
}
