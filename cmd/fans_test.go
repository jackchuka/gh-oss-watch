package cmd

import (
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
)

func TestAggregateFans(t *testing.T) {
	tests := []struct {
		name       string
		repoUsers  map[string][]services.UserAPIData
		wantFans   int
		wantStars  int
		wantFirst  string
		wantFirstN int
		wantOrder  []string
	}{
		{
			name: "aggregates across repos and sorts by count desc then repo asc",
			repoUsers: map[string][]services.UserAPIData{
				"owner/repo1": {{Login: "alice"}, {Login: "bob"}},
				"owner/repo2": {{Login: "alice"}, {Login: "charlie"}},
				"owner/repo3": {{Login: "alice"}},
			},
			wantFans:   3,
			wantStars:  5,
			wantFirst:  "alice",
			wantFirstN: 3,
		},
		{
			name: "same count fans sorted by first repo then login",
			repoUsers: map[string][]services.UserAPIData{
				"owner/zebra": {{Login: "dan"}},
				"owner/alpha": {{Login: "eve"}, {Login: "dan"}},
				"owner/beta":  {{Login: "frank"}},
			},
			wantFans:   3,
			wantStars:  4,
			wantFirst:  "dan",
			wantFirstN: 2,
			wantOrder:  []string{"dan", "eve", "frank"},
		},
		{
			name:       "empty input returns empty result",
			repoUsers:  map[string][]services.UserAPIData{},
			wantFans:   0,
			wantStars:  0,
			wantFirst:  "",
			wantFirstN: 0,
		},
		{
			name: "single user single repo",
			repoUsers: map[string][]services.UserAPIData{
				"owner/repo1": {{Login: "alice"}},
			},
			wantFans:   1,
			wantStars:  1,
			wantFirst:  "alice",
			wantFirstN: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := aggregateFans(tt.repoUsers)

			if result.TotalFans != tt.wantFans {
				t.Errorf("TotalFans = %d, want %d", result.TotalFans, tt.wantFans)
			}
			if result.TotalStars != tt.wantStars {
				t.Errorf("TotalStars = %d, want %d", result.TotalStars, tt.wantStars)
			}
			if tt.wantFirst != "" {
				if len(result.Fans) == 0 {
					t.Fatal("expected at least one fan")
				}
				if result.Fans[0].Login != tt.wantFirst {
					t.Errorf("first fan = %q, want %q", result.Fans[0].Login, tt.wantFirst)
				}
				if result.Fans[0].Count != tt.wantFirstN {
					t.Errorf("first fan count = %d, want %d", result.Fans[0].Count, tt.wantFirstN)
				}
			}
			for i, want := range tt.wantOrder {
				if i >= len(result.Fans) {
					t.Errorf("expected fan at index %d (%q), but only %d fans", i, want, len(result.Fans))
					break
				}
				if result.Fans[i].Login != want {
					t.Errorf("fan[%d] = %q, want %q", i, result.Fans[i].Login, want)
				}
			}
		})
	}
}
