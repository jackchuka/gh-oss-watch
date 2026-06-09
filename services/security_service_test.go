package services

import "testing"

func TestSeverityWeight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"critical", "critical", 4},
		{"high", "high", 3},
		{"medium", "medium", 2},
		{"low", "low", 1},
		{"case insensitive", "CRITICAL", 4},
		{"empty", "", 0},
		{"unknown", "bogus", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SeverityWeight(tt.input); got != tt.want {
				t.Errorf("SeverityWeight(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSecurityAlert(t *testing.T) {
	var raw DependabotAlertAPIData
	raw.SecurityAdvisory.GHSAID = "GHSA-xxxx"
	raw.SecurityVulnerability.Severity = "high"
	raw.SecurityVulnerability.VulnerableRange = "< 2.20.0"
	raw.SecurityVulnerability.FirstPatchedVersion.Identifier = "2.20.0"
	raw.SecurityVulnerability.Package.Name = "pygments"
	raw.SecurityVulnerability.Package.Ecosystem = "pip"
	raw.Dependency.Scope = "development"

	got := toSecurityAlert(raw)
	want := SecurityAlert{
		Severity: "high", Ecosystem: "pip", Package: "pygments",
		VulnRange: "< 2.20.0", FixedVersion: "2.20.0", GHSA: "GHSA-xxxx", Scope: "development",
	}
	if got != want {
		t.Errorf("toSecurityAlert() = %+v, want %+v", got, want)
	}
}
