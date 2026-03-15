package cmd

import (
	"errors"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleConfigAdd(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		events  []string
		setup   func(*mock_services.MockConfigService, *mock_services.MockGitHubService)
		wantErr bool
	}{
		{
			name:   "success",
			repo:   "microsoft/vscode",
			events: []string{"stars", "issues"},
			setup: func(mc *mock_services.MockConfigService, mg *mock_services.MockGitHubService) {
				config := &services.Config{Repos: []services.RepoConfig{}}
				mc.EXPECT().Load().Return(config, nil)
				mg.EXPECT().GetRepoInfo("microsoft", "vscode").Return(&services.RepoAPIData{Language: "TypeScript"}, nil)
				mc.EXPECT().Save(gomock.Any()).DoAndReturn(func(c *services.Config) error {
					if len(c.Repos) != 1 {
						t.Errorf("expected 1 repo, got %d", len(c.Repos))
					}
					if c.Repos[0].Repo != "microsoft/vscode" {
						t.Errorf("expected 'microsoft/vscode', got %s", c.Repos[0].Repo)
					}
					if c.Repos[0].Language != "TypeScript" {
						t.Errorf("expected language 'TypeScript', got %s", c.Repos[0].Language)
					}
					return nil
				})
			},
		},
		{
			name:   "repo does not exist",
			repo:   "notexisting/repo",
			events: []string{"invalid_event"},
			setup: func(mc *mock_services.MockConfigService, mg *mock_services.MockGitHubService) {
				config := &services.Config{Repos: []services.RepoConfig{}}
				mc.EXPECT().Load().Return(config, nil)
				mg.EXPECT().GetRepoInfo("notexisting", "repo").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:   "invalid events",
			repo:   "microsoft/vscode",
			events: []string{"invalid_event"},
			setup: func(mc *mock_services.MockConfigService, mg *mock_services.MockGitHubService) {
				config := &services.Config{Repos: []services.RepoConfig{}}
				mc.EXPECT().Load().Return(config, nil)
				mg.EXPECT().GetRepoInfo("microsoft", "vscode").Return(&services.RepoAPIData{}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)
			mockGitHub := mock_services.NewMockGitHubService(ctrl)
			tt.setup(mockConfig, mockGitHub)

			err := handleConfigAdd(mockConfig, mockGitHub, tt.repo, tt.events)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
