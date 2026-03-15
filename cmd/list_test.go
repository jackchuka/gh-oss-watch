package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleList(t *testing.T) {
	tests := []struct {
		name    string
		lang    string
		format  string
		setup   func(*mock_services.MockConfigService)
		check   func(t *testing.T, out string)
		wantErr bool
	}{
		{
			name:   "lists all repos",
			lang:   "",
			format: "plain",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "facebook/react", Events: []string{"stars", "issues"}, Language: "JavaScript"},
					{Repo: "golang/go", Events: []string{"stars", "forks"}, Language: "Go"},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				for _, s := range []string{"facebook/react", "JavaScript", "golang/go", "Go"} {
					if !contains(out, s) {
						t.Errorf("expected output to contain %q, got:\n%s", s, out)
					}
				}
			},
		},
		{
			name:   "filters by language case-insensitive",
			lang:   "go",
			format: "plain",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "facebook/react", Events: []string{"stars"}, Language: "JavaScript"},
					{Repo: "golang/go", Events: []string{"stars"}, Language: "Go"},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				if !contains(out, "golang/go") {
					t.Errorf("expected output to contain 'golang/go', got:\n%s", out)
				}
				if contains(out, "facebook/react") {
					t.Errorf("expected output NOT to contain 'facebook/react', got:\n%s", out)
				}
			},
		},
		{
			name:   "unknown lang returns empty list",
			lang:   "rust",
			format: "plain",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "golang/go", Events: []string{"stars"}, Language: "Go"},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				if contains(out, "golang/go") {
					t.Errorf("expected empty output, got:\n%s", out)
				}
			},
		},
		{
			name:   "empty config",
			lang:   "",
			format: "plain",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				if contains(out, "REPO") {
					t.Errorf("expected no table header for empty config, got:\n%s", out)
				}
			},
		},
		{
			name:   "json output with wrapper",
			lang:   "",
			format: "json",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "golang/go", Events: []string{"stars"}, Language: "Go"},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
				}
				if _, ok := result["repos"]; !ok {
					t.Error("expected 'repos' key in JSON output")
				}
				if total, ok := result["total"]; !ok || total != float64(1) {
					t.Errorf("expected total=1, got %v", total)
				}
			},
		},
		{
			name:   "json output with lang filter",
			lang:   "go",
			format: "json",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "facebook/react", Events: []string{"stars"}, Language: "JavaScript"},
					{Repo: "golang/go", Events: []string{"stars"}, Language: "Go"},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(out), &result); err != nil {
					t.Fatalf("output is not valid JSON: %v", err)
				}
				if total := result["total"]; total != float64(1) {
					t.Errorf("expected total=1 after filter, got %v", total)
				}
			},
		},
		{
			name:   "repo with no language shows dash in plain",
			lang:   "",
			format: "plain",
			setup: func(mc *mock_services.MockConfigService) {
				config := &services.Config{Repos: []services.RepoConfig{
					{Repo: "owner/config-only", Events: []string{"stars"}, Language: ""},
				}}
				mc.EXPECT().Load().Return(config, nil)
			},
			check: func(t *testing.T, out string) {
				if !contains(out, "-") {
					t.Errorf("expected '-' for empty language, got:\n%s", out)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockConfig := mock_services.NewMockConfigService(ctrl)
			tt.setup(mockConfig)

			var buf bytes.Buffer
			var formatter services.Formatter
			if tt.format == "json" {
				formatter = services.NewJSONFormatter(&buf)
			} else {
				formatter = services.NewPlainFormatter(&buf, false, 120)
			}

			err := handleList(mockConfig, formatter, tt.lang)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, buf.String())
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && bytes.Contains([]byte(s), []byte(substr))
}
