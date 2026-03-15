package services

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestJSONFormatter_RenderStatus(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		entries []StatusEntry
		check   func(t *testing.T, got []map[string]any)
	}{
		{
			name: "with entries produces valid JSON array with correct fields",
			entries: []StatusEntry{
				{
					EventSummary: EventSummary{
						Repo:            "owner/repo",
						NewStars:        5,
						NewIssues:       2,
						NewPRs:          1,
						NewForks:        3,
						HasChanges:      true,
						NewRelease:      false,
						UnreleasedCount: 4,
						ReleaseTag:      "v1.0.0",
					},
					TotalStars:  100,
					TotalIssues: 10,
					TotalPRs:    7,
					TotalForks:  20,
				},
			},
			check: func(t *testing.T, got []map[string]any) {
				if len(got) != 1 {
					t.Fatalf("expected 1 entry, got %d", len(got))
				}
				entry := got[0]
				assertEqual(t, "owner/repo", entry["repo"])
				assertEqual(t, float64(5), entry["newStars"])
				assertEqual(t, float64(2), entry["newIssues"])
				assertEqual(t, float64(1), entry["newPRs"])
				assertEqual(t, float64(3), entry["newForks"])
				assertEqual(t, true, entry["hasChanges"])
				assertEqual(t, false, entry["newRelease"])
				assertEqual(t, float64(4), entry["unreleasedCount"])
				assertEqual(t, "v1.0.0", entry["releaseTag"])
				assertEqual(t, float64(100), entry["totalStars"])
				assertEqual(t, float64(10), entry["totalIssues"])
				assertEqual(t, float64(7), entry["totalPRs"])
				assertEqual(t, float64(20), entry["totalForks"])
				// Events field is tagged json:"-" and should not appear
				if _, ok := entry["Events"]; ok {
					t.Error("Events field should not be present in JSON output")
				}
			},
		},
		{
			name:    "nil entries produces empty JSON array not null",
			entries: nil,
			check: func(t *testing.T, got []map[string]any) {
				if got == nil {
					t.Fatal("expected empty array, got null")
				}
				if len(got) != 0 {
					t.Fatalf("expected 0 entries, got %d", len(got))
				}
			},
		},
		{
			name:    "empty slice produces empty JSON array",
			entries: []StatusEntry{},
			check: func(t *testing.T, got []map[string]any) {
				if got == nil {
					t.Fatal("expected empty array, got null")
				}
				if len(got) != 0 {
					t.Fatalf("expected 0 entries, got %d", len(got))
				}
			},
		},
	}

	_ = now // used by future tests if needed

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewJSONFormatter(&buf)

			err := f.RenderStatus(tt.entries)
			if err != nil {
				t.Fatalf("RenderStatus returned error: %v", err)
			}

			// Verify raw output for nil case
			if tt.entries == nil {
				raw := buf.String()
				if raw != "[]\n" {
					t.Errorf("expected `[]\\n`, got %q", raw)
				}
			}

			var got []map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
			}

			tt.check(t, got)
		})
	}
}

func TestJSONFormatter_RenderDashboard(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	result := DashboardResult{
		Repos: []DashboardEntry{
			{
				Repo:            "owner/repo",
				Stars:           100,
				Issues:          5,
				PullRequests:    3,
				Forks:           20,
				UpdatedAt:       now,
				LatestRelease:   "v1.2.0",
				UnreleasedCount: 2,
				Watching:        []string{"stars", "issues"},
			},
		},
		Totals: DashboardTotals{
			Stars:       100,
			Issues:      5,
			PRs:         3,
			Forks:       20,
			NeedRelease: 1,
		},
	}

	var buf bytes.Buffer
	f := NewJSONFormatter(&buf)

	err := f.RenderDashboard(result)
	if err != nil {
		t.Fatalf("RenderDashboard returned error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	// Verify top-level keys
	if _, ok := got["repos"]; !ok {
		t.Error("expected 'repos' key in output")
	}
	if _, ok := got["totals"]; !ok {
		t.Error("expected 'totals' key in output")
	}

	repos, ok := got["repos"].([]any)
	if !ok {
		t.Fatalf("'repos' should be an array, got %T", got["repos"])
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	repo := repos[0].(map[string]any)
	assertEqual(t, "owner/repo", repo["repo"])
	assertEqual(t, float64(100), repo["stars"])
	assertEqual(t, float64(5), repo["issues"])
	assertEqual(t, float64(3), repo["pullRequests"])
	assertEqual(t, float64(20), repo["forks"])
	assertEqual(t, "v1.2.0", repo["latestRelease"])
	assertEqual(t, float64(2), repo["unreleasedCount"])

	watching, ok := repo["watching"].([]any)
	if !ok {
		t.Fatalf("'watching' should be an array, got %T", repo["watching"])
	}
	if len(watching) != 2 {
		t.Fatalf("expected 2 watching entries, got %d", len(watching))
	}

	totals := got["totals"].(map[string]any)
	assertEqual(t, float64(100), totals["stars"])
	assertEqual(t, float64(5), totals["issues"])
	assertEqual(t, float64(3), totals["pullRequests"])
	assertEqual(t, float64(20), totals["forks"])
	assertEqual(t, float64(1), totals["needRelease"])
}

func TestJSONFormatter_RenderReleases(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		releases []ReleaseInfo
		check    func(t *testing.T, got []map[string]any)
	}{
		{
			name: "with releases produces valid JSON array with correct fields",
			releases: []ReleaseInfo{
				{
					Repo:            "owner/repo",
					LatestRelease:   "v1.0.0",
					ReleaseDate:     now,
					ReleaseAge:      "30 days ago",
					UnreleasedCount: 5,
					Status:          "behind",
				},
			},
			check: func(t *testing.T, got []map[string]any) {
				if len(got) != 1 {
					t.Fatalf("expected 1 release, got %d", len(got))
				}
				r := got[0]
				assertEqual(t, "owner/repo", r["repo"])
				assertEqual(t, "v1.0.0", r["latestRelease"])
				assertEqual(t, "30 days ago", r["releaseAge"])
				assertEqual(t, float64(5), r["unreleasedCount"])
				assertEqual(t, "behind", r["status"])
				if _, ok := r["releaseDate"]; !ok {
					t.Error("expected 'releaseDate' field in output")
				}
			},
		},
		{
			name:     "nil releases produces empty JSON array not null",
			releases: nil,
			check: func(t *testing.T, got []map[string]any) {
				if got == nil {
					t.Fatal("expected empty array, got null")
				}
				if len(got) != 0 {
					t.Fatalf("expected 0 releases, got %d", len(got))
				}
			},
		},
		{
			name:     "empty slice produces empty JSON array",
			releases: []ReleaseInfo{},
			check: func(t *testing.T, got []map[string]any) {
				if got == nil {
					t.Fatal("expected empty array, got null")
				}
				if len(got) != 0 {
					t.Fatalf("expected 0 releases, got %d", len(got))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewJSONFormatter(&buf)

			err := f.RenderReleases(tt.releases)
			if err != nil {
				t.Fatalf("RenderReleases returned error: %v", err)
			}

			// Verify raw output for nil case
			if tt.releases == nil {
				raw := buf.String()
				if raw != "[]\n" {
					t.Errorf("expected `[]\\n`, got %q", raw)
				}
			}

			var got []map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
			}

			tt.check(t, got)
		})
	}
}

func TestJSONFormatter_RenderFans(t *testing.T) {
	tests := []struct {
		name   string
		result FansResult
		check  func(t *testing.T, got map[string]any)
	}{
		{
			name: "with fans produces valid JSON with correct fields",
			result: FansResult{
				Fans: []FanEntry{
					{Login: "userA", Count: 3, Repos: []string{"owner/repo1", "owner/repo2", "owner/repo3"}},
					{Login: "userB", Count: 1, Repos: []string{"owner/repo1"}},
				},
				TotalFans:  2,
				TotalStars: 4,
			},
			check: func(t *testing.T, got map[string]any) {
				fans, ok := got["fans"].([]any)
				if !ok {
					t.Fatalf("'fans' should be an array, got %T", got["fans"])
				}
				if len(fans) != 2 {
					t.Fatalf("expected 2 fans, got %d", len(fans))
				}
				fan := fans[0].(map[string]any)
				assertEqual(t, "userA", fan["login"])
				assertEqual(t, float64(3), fan["count"])
				repos, ok := fan["repos"].([]any)
				if !ok {
					t.Fatalf("'repos' should be an array, got %T", fan["repos"])
				}
				if len(repos) != 3 {
					t.Fatalf("expected 3 repos, got %d", len(repos))
				}
				assertEqual(t, float64(2), got["totalFans"])
				assertEqual(t, float64(4), got["totalStars"])
			},
		},
		{
			name: "empty fans produces empty array not null",
			result: FansResult{
				Fans:       []FanEntry{},
				TotalFans:  0,
				TotalStars: 0,
			},
			check: func(t *testing.T, got map[string]any) {
				fans, ok := got["fans"].([]any)
				if !ok {
					t.Fatalf("'fans' should be an array, got %T", got["fans"])
				}
				if len(fans) != 0 {
					t.Fatalf("expected 0 fans, got %d", len(fans))
				}
			},
		},
		{
			name:   "nil fans produces empty array not null",
			result: FansResult{},
			check: func(t *testing.T, got map[string]any) {
				fans, ok := got["fans"].([]any)
				if !ok {
					t.Fatalf("'fans' should be an array, got %T", got["fans"])
				}
				if len(fans) != 0 {
					t.Fatalf("expected 0 fans, got %d", len(fans))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewJSONFormatter(&buf)

			err := f.RenderFans(tt.result)
			if err != nil {
				t.Fatalf("RenderFans returned error: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
			}

			tt.check(t, got)
		})
	}
}

func TestJSONFormatter_RenderList(t *testing.T) {
	tests := []struct {
		name   string
		result ListResult
		check  func(t *testing.T, got map[string]any)
	}{
		{
			name: "with repos produces valid JSON with repos and total",
			result: ListResult{
				Repos: []RepoConfig{
					{Repo: "golang/go", Language: "Go", Events: []string{"stars"}},
				},
				Total: 1,
			},
			check: func(t *testing.T, got map[string]any) {
				repos, ok := got["repos"].([]any)
				if !ok {
					t.Fatalf("'repos' should be an array, got %T", got["repos"])
				}
				if len(repos) != 1 {
					t.Fatalf("expected 1 repo, got %d", len(repos))
				}
				repo := repos[0].(map[string]any)
				assertEqual(t, "golang/go", repo["repo"])
				assertEqual(t, "Go", repo["language"])
				assertEqual(t, float64(1), got["total"])
			},
		},
		{
			name: "empty repos produces empty array not null",
			result: ListResult{
				Repos: []RepoConfig{},
				Total: 0,
			},
			check: func(t *testing.T, got map[string]any) {
				repos, ok := got["repos"].([]any)
				if !ok {
					t.Fatalf("'repos' should be an array, got %T", got["repos"])
				}
				if len(repos) != 0 {
					t.Fatalf("expected 0 repos, got %d", len(repos))
				}
				assertEqual(t, float64(0), got["total"])
			},
		},
		{
			name:   "nil repos produces empty array not null",
			result: ListResult{},
			check: func(t *testing.T, got map[string]any) {
				repos, ok := got["repos"].([]any)
				if !ok {
					t.Fatalf("'repos' should be an array, got %T", got["repos"])
				}
				if len(repos) != 0 {
					t.Fatalf("expected 0 repos, got %d", len(repos))
				}
			},
		},
		{
			name: "empty language renders as empty string not omitted",
			result: ListResult{
				Repos: []RepoConfig{
					{Repo: "owner/repo", Language: "", Events: []string{"stars"}},
				},
				Total: 1,
			},
			check: func(t *testing.T, got map[string]any) {
				repos := got["repos"].([]any)
				repo := repos[0].(map[string]any)
				lang, ok := repo["language"]
				if !ok {
					t.Fatal("expected 'language' key to be present in JSON output")
				}
				assertEqual(t, "", lang)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewJSONFormatter(&buf)

			err := f.RenderList(tt.result)
			if err != nil {
				t.Fatalf("RenderList returned error: %v", err)
			}

			var got map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
			}

			tt.check(t, got)
		})
	}
}

func assertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v (%T), got %v (%T)", expected, expected, actual, actual)
	}
}
