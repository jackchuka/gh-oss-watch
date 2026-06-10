package services

import (
	"errors"
	"sort"
	"strings"
)

func SeverityWeight(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func toSecurityAlert(a DependabotAlertAPIData) SecurityAlert {
	return SecurityAlert{
		Severity:     a.SecurityVulnerability.Severity,
		Ecosystem:    a.SecurityVulnerability.Package.Ecosystem,
		Package:      a.SecurityVulnerability.Package.Name,
		VulnRange:    a.SecurityVulnerability.VulnerableRange,
		FixedVersion: a.SecurityVulnerability.FirstPatchedVersion.Identifier,
		GHSA:         a.SecurityAdvisory.GHSAID,
		Scope:        a.Dependency.Scope,
	}
}

// BuildSecurityResult aggregates per-repo alerts into a ranked snapshot.
// errs[i] == ErrNoAlertAccess routes a repo to SkippedRepos; any other non-nil
// error excludes the repo silently (the caller is responsible for logging it).
// severityFloor (""|low|medium|high|critical) drops alerts below that severity.
func BuildSecurityResult(repos []string, alertsPerRepo [][]SecurityAlert, errs []error, severityFloor string) SecurityResult {
	floor := SeverityWeight(severityFloor)
	result := SecurityResult{
		Totals:       map[string]int{},
		WatchedCount: len(repos),
	}

	for i, repo := range repos {
		if i < len(errs) && errs[i] != nil {
			if errors.Is(errs[i], ErrNoAlertAccess) {
				result.SkippedRepos = append(result.SkippedRepos, repo)
			}
			continue
		}

		counts := map[string]int{}
		var filtered []SecurityAlert
		for _, a := range alertsPerRepo[i] {
			if SeverityWeight(a.Severity) < floor {
				continue
			}
			sev := strings.ToLower(a.Severity)
			filtered = append(filtered, a)
			counts[sev]++
			result.Totals[sev]++
			result.GrandTotal++
		}

		if len(filtered) == 0 {
			continue
		}

		sortAlerts(filtered)
		result.Repos = append(result.Repos, SecurityRepoEntry{
			Repo:   repo,
			Total:  len(filtered),
			Counts: counts,
			Alerts: filtered,
		})
	}

	sortSecurityRepos(result.Repos)
	return result
}

func sortAlerts(alerts []SecurityAlert) {
	sort.SliceStable(alerts, func(i, j int) bool {
		wi, wj := SeverityWeight(alerts[i].Severity), SeverityWeight(alerts[j].Severity)
		if wi != wj {
			return wi > wj
		}
		return alerts[i].Package < alerts[j].Package
	})
}

func sortSecurityRepos(repos []SecurityRepoEntry) {
	sort.SliceStable(repos, func(i, j int) bool {
		for _, sev := range []string{"critical", "high", "medium", "low"} {
			if repos[i].Counts[sev] != repos[j].Counts[sev] {
				return repos[i].Counts[sev] > repos[j].Counts[sev]
			}
		}
		if repos[i].Total != repos[j].Total {
			return repos[i].Total > repos[j].Total
		}
		return repos[i].Repo < repos[j].Repo
	})
}
