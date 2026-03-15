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
