package cmd

import (
	"fmt"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleConfigSet(t *testing.T) {
	tests := []struct {
		name    string
		repo    string
		events  []string
		setup   func(*mock_services.MockConfigService)
		wantErr bool
	}{
		{
			name:   "success",
			repo:   "microsoft/vscode",
			events: []string{"stars", "forks"},
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "microsoft/vscode", Events: []string{"stars"}},
				}}
				mc.EXPECT().Load().Return(config, nil)
				mc.EXPECT().Save(gomock.Any()).Return(nil)
			},
		},
		{
			name:   "repo not in config",
			repo:   "unknown/repo",
			events: []string{"stars"},
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{}}
				mc.EXPECT().Load().Return(config, nil)
			},
			wantErr: true,
		},
		{
			name:    "no events",
			repo:    "microsoft/vscode",
			events:  []string{},
			setup:   func(mc *mock_services.MockConfigService) {},
			wantErr: true,
		},
		{
			name:   "load error",
			repo:   "microsoft/vscode",
			events: []string{"stars"},
			setup: func(mc *mock_services.MockConfigService) {
				mc.EXPECT().Load().Return(nil, fmt.Errorf("load failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)
			tt.setup(mockConfig)

			err := handleConfigSet(mockConfig, tt.repo, tt.events)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
