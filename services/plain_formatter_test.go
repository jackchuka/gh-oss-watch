package services

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestPlainFormatter_RenderStatus(t *testing.T) {
	tests := []struct {
		name     string
		entries  []StatusEntry
		contains []string
	}{
		{
			name: "with entries contains repo name, diffs, and section header",
			entries: []StatusEntry{
				{
					EventSummary: EventSummary{
						Repo:       "owner/repo",
						NewStars:   5,
						NewIssues:  2,
						NewPRs:     1,
						NewForks:   3,
						HasChanges: true,
					},
					Events: []string{"stars", "issues", "pull_requests", "forks"},
				},
			},
			contains: []string{"owner/repo", "+5", "+2", "+1", "+3", "📈"},
		},
		{
			name: "with release entry shows release tag",
			entries: []StatusEntry{
				{
					EventSummary: EventSummary{
						Repo:       "owner/repo",
						NewRelease: true,
						ReleaseTag: "v1.2.0",
					},
					Events: []string{"releases"},
				},
			},
			contains: []string{"owner/repo", "v1.2.0"},
		},
		{
			name: "with unreleased commits shows unreleased count",
			entries: []StatusEntry{
				{
					EventSummary: EventSummary{
						Repo:            "owner/repo",
						UnreleasedCount: 4,
					},
					Events: []string{"releases"},
				},
			},
			contains: []string{"owner/repo", "4 unreleased"},
		},
		{
			name:     "nil entries shows no activity message",
			entries:  nil,
			contains: []string{"No new activity"},
		},
		{
			name:     "empty entries shows no activity message",
			entries:  []StatusEntry{},
			contains: []string{"No new activity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewPlainFormatter(&buf, false, 120)

			err := f.RenderStatus(tt.entries)
			if err != nil {
				t.Fatalf("RenderStatus returned error: %v", err)
			}

			out := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, out)
				}
			}
		})
	}
}

func TestPlainFormatter_RenderDashboard(t *testing.T) {
	now := time.Now().Add(-48 * time.Hour)

	result := DashboardResult{
		Repos: []DashboardEntry{
			{
				Repo:          "owner/repo",
				Stars:         100,
				Issues:        5,
				PullRequests:  3,
				Forks:         20,
				UpdatedAt:     now,
				LatestRelease: "v1.2.0",
			},
		},
		Totals: DashboardTotals{
			Stars:  100,
			Issues: 5,
			PRs:    3,
			Forks:  20,
		},
	}

	var buf bytes.Buffer
	f := NewPlainFormatter(&buf, false, 120)

	err := f.RenderDashboard(result)
	if err != nil {
		t.Fatalf("RenderDashboard returned error: %v", err)
	}

	out := buf.String()
	for _, s := range []string{"owner/repo", "100", "📊", "Totals"} {
		if !strings.Contains(out, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, out)
		}
	}
}

func TestPlainFormatter_RenderFans(t *testing.T) {
	tests := []struct {
		name     string
		result   FansResult
		contains []string
	}{
		{
			name: "with fans shows table with header and summary",
			result: FansResult{
				Fans: []FanEntry{
					{Login: "userA", Count: 3, Repos: []string{"owner/repo1", "owner/repo2", "owner/repo3"}},
					{Login: "userB", Count: 1, Repos: []string{"owner/repo1"}},
				},
				TotalFans:  2,
				TotalStars: 4,
			},
			contains: []string{"userA", "userB", "3", "1", "🌟", "2 fans", "4 stars"},
		},
		{
			name: "empty fans shows no fans message",
			result: FansResult{
				Fans:       []FanEntry{},
				TotalFans:  0,
				TotalStars: 0,
			},
			contains: []string{"No stargazers found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewPlainFormatter(&buf, false, 120)

			err := f.RenderFans(tt.result)
			if err != nil {
				t.Fatalf("RenderFans returned error: %v", err)
			}

			out := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, out)
				}
			}
		})
	}
}

func TestPlainFormatter_RenderList(t *testing.T) {
	tests := []struct {
		name     string
		result   ListResult
		contains []string
	}{
		{
			name: "with repos shows table with language and events",
			result: ListResult{
				Repos: []RepoConfig{
					{Repo: "facebook/react", Language: "JavaScript", Events: []string{"stars", "issues"}},
					{Repo: "golang/go", Language: "Go", Events: []string{"stars", "forks"}},
				},
				Total: 2,
			},
			contains: []string{"facebook/react", "JavaScript", "golang/go", "Go", "stars, issues", "stars, forks"},
		},
		{
			name: "empty language shows dash",
			result: ListResult{
				Repos: []RepoConfig{
					{Repo: "owner/repo", Language: "", Events: []string{"stars"}},
				},
				Total: 1,
			},
			contains: []string{"owner/repo", "-"},
		},
		{
			name: "empty repos shows no repos message",
			result: ListResult{
				Repos: []RepoConfig{},
				Total: 0,
			},
			contains: []string{"No repos"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewPlainFormatter(&buf, false, 120)

			err := f.RenderList(tt.result)
			if err != nil {
				t.Fatalf("RenderList returned error: %v", err)
			}

			out := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("expected output to contain %q, got:\n%s", s, out)
				}
			}
		})
	}
}

func TestPlainRenderSecurity_AllClear(t *testing.T) {
	var buf bytes.Buffer
	f := NewPlainFormatter(&buf, false, 80)
	if err := f.RenderSecurity(SecurityResult{WatchedCount: 5}, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No open security alerts across 5 watched repos") {
		t.Errorf("missing all-clear message: %q", buf.String())
	}
}

func TestPlainRenderSecurity_TableAndSkipped(t *testing.T) {
	var buf bytes.Buffer
	f := NewPlainFormatter(&buf, false, 120)
	res := SecurityResult{
		Repos: []SecurityRepoEntry{
			{Repo: "o/a", Total: 2, Counts: map[string]int{"critical": 1, "low": 1},
				Alerts: []SecurityAlert{{Severity: "critical", Package: "log4j"}, {Severity: "low", Package: "zlib"}}},
		},
		Totals: map[string]int{"critical": 1, "low": 1}, GrandTotal: 2,
		WatchedCount: 3, SkippedRepos: []string{"o/c"},
	}
	if err := f.RenderSecurity(res, false); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "o/a") {
		t.Errorf("missing repo row: %q", out)
	}
	if !strings.Contains(out, "1 watched repo(s) skipped") {
		t.Errorf("missing skipped footer: %q", out)
	}
}

func TestPlainRenderSecurity_Detail(t *testing.T) {
	var buf bytes.Buffer
	f := NewPlainFormatter(&buf, false, 120)
	res := SecurityResult{
		Repos: []SecurityRepoEntry{
			{Repo: "o/a", Total: 1, Counts: map[string]int{"high": 1},
				Alerts: []SecurityAlert{{Severity: "high", Ecosystem: "pip", Package: "pygments",
					VulnRange: "< 2.20.0", FixedVersion: "2.20.0", GHSA: "GHSA-x", Scope: "development"}}},
		},
		Totals: map[string]int{"high": 1}, GrandTotal: 1, WatchedCount: 1,
	}
	if err := f.RenderSecurity(res, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "pip/pygments < 2.20.0 → 2.20.0") {
		t.Errorf("missing detail line: %q", buf.String())
	}
}

func TestPlainFormatter_RenderReleases(t *testing.T) {
	releases := []ReleaseInfo{
		{
			Repo:            "owner/repo",
			LatestRelease:   "v1.0.0",
			ReleaseAge:      "30 days",
			UnreleasedCount: 5,
			Status:          "needs release",
		},
		{
			Repo:          "owner/other",
			LatestRelease: "v2.0.0",
			ReleaseAge:    "5 days",
			Status:        "up to date",
		},
	}

	var buf bytes.Buffer
	f := NewPlainFormatter(&buf, false, 120)

	err := f.RenderReleases(releases)
	if err != nil {
		t.Fatalf("RenderReleases returned error: %v", err)
	}

	out := buf.String()
	for _, s := range []string{"owner/repo", "v1.0.0", "needs release", "📦", "repos need a release", "up to date"} {
		if !strings.Contains(out, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, out)
		}
	}
}
