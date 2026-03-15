package cmd

import (
	"fmt"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleInit(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mock_services.MockConfigService)
		wantErr   bool
		errString string
	}{
		{
			name: "success",
			setup: func(m *mock_services.MockConfigService) {
				m.EXPECT().Load().Return(&services.Config{Repos: []services.RepoConfig{}}, nil)
				m.EXPECT().GetConfigPath().Return("/mock/config.yaml", nil)
				m.EXPECT().Save(gomock.Any()).Return(nil)
			},
		},
		{
			name: "load error",
			setup: func(m *mock_services.MockConfigService) {
				m.EXPECT().Load().Return(nil, fmt.Errorf("load failed"))
			},
			wantErr:   true,
			errString: "load failed",
		},
		{
			name: "save error",
			setup: func(m *mock_services.MockConfigService) {
				m.EXPECT().Load().Return(&services.Config{Repos: []services.RepoConfig{}}, nil)
				m.EXPECT().GetConfigPath().Return("/mock/config.yaml", nil)
				m.EXPECT().Save(gomock.Any()).Return(fmt.Errorf("save failed"))
			},
			wantErr:   true,
			errString: "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)
			tt.setup(mockConfig)

			err := handleInit(mockConfig)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.errString {
					t.Errorf("error = %q, want %q", err.Error(), tt.errString)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
