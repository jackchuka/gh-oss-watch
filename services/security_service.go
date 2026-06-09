package services

import "strings"

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
