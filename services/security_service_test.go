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

func sampleAlerts() [][]SecurityAlert {
	return [][]SecurityAlert{
		{ // repo 0: one high, one low
			{Severity: "low", Package: "zlib"},
			{Severity: "high", Package: "openssl"},
		},
		{ // repo 1: one critical
			{Severity: "critical", Package: "log4j"},
		},
		nil, // repo 2: no access (see errs)
	}
}

func TestBuildSecurityResult_RankingAndTotals(t *testing.T) {
	repos := []string{"o/a", "o/b", "o/c"}
	errs := []error{nil, nil, ErrNoAlertAccess}

	res := BuildSecurityResult(repos, sampleAlerts(), errs, "")

	if res.WatchedCount != 3 {
		t.Errorf("WatchedCount = %d, want 3", res.WatchedCount)
	}
	if res.GrandTotal != 3 {
		t.Errorf("GrandTotal = %d, want 3", res.GrandTotal)
	}
	if len(res.Repos) != 2 {
		t.Fatalf("len(Repos) = %d, want 2", len(res.Repos))
	}
	if res.Repos[0].Repo != "o/b" {
		t.Errorf("Repos[0] = %s, want o/b", res.Repos[0].Repo)
	}
	if res.Repos[1].Alerts[0].Severity != "high" {
		t.Errorf("o/a first alert = %s, want high", res.Repos[1].Alerts[0].Severity)
	}
	if res.Totals["critical"] != 1 || res.Totals["high"] != 1 || res.Totals["low"] != 1 {
		t.Errorf("Totals = %+v, want crit1 high1 low1", res.Totals)
	}
	if len(res.SkippedRepos) != 1 || res.SkippedRepos[0] != "o/c" {
		t.Errorf("SkippedRepos = %v, want [o/c]", res.SkippedRepos)
	}
}

func TestBuildSecurityResult_SeverityFloor(t *testing.T) {
	repos := []string{"o/a", "o/b", "o/c"}
	errs := []error{nil, nil, ErrNoAlertAccess}

	res := BuildSecurityResult(repos, sampleAlerts(), errs, "high")

	if res.GrandTotal != 2 {
		t.Errorf("GrandTotal = %d, want 2", res.GrandTotal)
	}
	if len(res.Repos) != 2 {
		t.Fatalf("len(Repos) = %d, want 2 (o/a keeps its high, o/b keeps its critical)", len(res.Repos))
	}
	for _, e := range res.Repos {
		for _, a := range e.Alerts {
			if SeverityWeight(a.Severity) < SeverityWeight("high") {
				t.Errorf("alert below floor leaked: %+v", a)
			}
		}
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
