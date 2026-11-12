package cmd

import (
	"slices"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleTag(t *testing.T) {
	newConfig := func() *services.Config {
		return &services.Config{Repos: []services.RepoConfig{
			{Repo: "acme/api", Events: []string{"stars"}, Tags: []string{"work"}},
			{Repo: "acme/web", Events: []string{"stars"}},
		}}
	}

	tests := []struct {
		name     string
		tag      string
		repos    []string
		add      bool
		wantSave bool
		wantErr  bool
		wantTags map[string][]string
	}{
		{
			name:     "tags several repos, normalizing case and skipping duplicates",
			tag:      "Work",
			repos:    []string{"acme/api", "acme/web"},
			add:      true,
			wantSave: true,
			wantTags: map[string][]string{"acme/api": {"work"}, "acme/web": {"work"}},
		},
		{
			name:     "untags a repo",
			tag:      "work",
			repos:    []string{"acme/api"},
			wantSave: true,
			wantTags: map[string][]string{"acme/api": nil},
		},
		{
			name:    "unknown repo leaves config untouched",
			tag:     "work",
			repos:   []string{"acme/web", "acme/missing"},
			add:     true,
			wantErr: true,
		},
		{
			name:    "invalid tag",
			tag:     "a b",
			repos:   []string{"acme/api"},
			add:     true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)
			config := newConfig()
			mockConfig.EXPECT().Load().Return(config, nil).AnyTimes()
			if tt.wantSave {
				mockConfig.EXPECT().Save(gomock.Any()).Return(nil)
			}

			err := handleTag(mockConfig, tt.tag, tt.repos, tt.add)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			for repo, want := range tt.wantTags {
				if got := config.GetRepo(repo).Tags; !slices.Equal(got, want) {
					t.Errorf("%s tags = %v, want %v", repo, got, want)
				}
			}
		})
	}
}

func TestHandleConfigAddWithTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockConfig := mock_services.NewMockConfigService(ctrl)
	mockGitHub := mock_services.NewMockGitHubService(ctrl)

	config := &services.Config{}
	mockConfig.EXPECT().Load().Return(config, nil)
	mockGitHub.EXPECT().GetRepoInfo("acme", "api").Return(&services.RepoAPIData{Language: "Go"}, nil)
	mockConfig.EXPECT().Save(gomock.Any()).Return(nil)

	if err := handleConfigAdd(mockConfig, mockGitHub, "acme/api", nil, []string{"Work", "oss"}); err != nil {
		t.Fatal(err)
	}
	if got := config.GetRepo("acme/api").Tags; !slices.Equal(got, []string{"work", "oss"}) {
		t.Errorf("tags = %v, want [work oss]", got)
	}
}
