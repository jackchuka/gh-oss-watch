package cmd

import (
	"fmt"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestBackfillLanguages(t *testing.T) {
	tests := []struct {
		name          string
		config        *services.Config
		repos         []services.RepoConfig
		stats         []*services.RepoStats
		expectSave    bool
		saveShouldErr bool
		wantLangs     []string
	}{
		{
			name: "backfills empty language from stats",
			config: &services.Config{Repos: []services.RepoConfig{
				{Repo: "owner/repo", Events: []string{"stars"}, Language: ""},
			}},
			stats:      []*services.RepoStats{{Language: "Go"}},
			expectSave: true,
			wantLangs:  []string{"Go"},
		},
		{
			name: "maps filtered stats back to the full config by repo name",
			config: &services.Config{Repos: []services.RepoConfig{
				{Repo: "owner/a", Events: []string{"stars"}},
				{Repo: "owner/b", Events: []string{"stars"}, Tags: []string{"work"}},
			}},
			repos:      []services.RepoConfig{{Repo: "owner/b", Events: []string{"stars"}, Tags: []string{"work"}}},
			stats:      []*services.RepoStats{{Language: "Rust"}},
			expectSave: true,
			wantLangs:  []string{"", "Rust"},
		},
		{
			name: "skips repos that already have language",
			config: &services.Config{Repos: []services.RepoConfig{
				{Repo: "owner/repo", Events: []string{"stars"}, Language: "Go"},
			}},
			stats:      []*services.RepoStats{{Language: "Go"}},
			expectSave: false,
			wantLangs:  []string{"Go"},
		},
		{
			name: "skips nil stats",
			config: &services.Config{Repos: []services.RepoConfig{
				{Repo: "owner/repo", Events: []string{"stars"}, Language: ""},
			}},
			stats:      []*services.RepoStats{nil},
			expectSave: false,
			wantLangs:  []string{""},
		},
		{
			name: "save failure does not panic",
			config: &services.Config{Repos: []services.RepoConfig{
				{Repo: "owner/repo", Events: []string{"stars"}, Language: ""},
			}},
			stats:         []*services.RepoStats{{Language: "Go"}},
			expectSave:    true,
			saveShouldErr: true,
			wantLangs:     []string{"Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)

			if tt.expectSave {
				if tt.saveShouldErr {
					mockConfig.EXPECT().Save(gomock.Any()).Return(fmt.Errorf("save failed"))
				} else {
					mockConfig.EXPECT().Save(gomock.Any()).Return(nil)
				}
			}

			repos := tt.repos
			if repos == nil {
				repos = tt.config.Repos
			}
			backfillLanguages(mockConfig, tt.config, repos, tt.stats)

			for i, want := range tt.wantLangs {
				if tt.config.Repos[i].Language != want {
					t.Errorf("repo %d: expected language %q, got %q", i, want, tt.config.Repos[i].Language)
				}
			}
		})
	}
}
