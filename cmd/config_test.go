package cmd

import (
	"errors"
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
	mock_services "github.com/jackchuka/gh-oss-watch/services/mock"
	"go.uber.org/mock/gomock"
)

func TestHandleConfigAdd_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockConfig := mock_services.NewMockConfigService(ctrl)
	mockCache := mock_services.NewMockCacheService(ctrl)
	mockGitHub := mock_services.NewMockGitHubService(ctrl)
	mockOutput := mock_services.NewMockOutput(ctrl)

	cli := NewCLI(mockConfig, mockCache, mockGitHub, mockOutput)

	config := &services.Config{Repos: []services.RepoConfig{}}

	//testing state
	repo_str := "microsoft/vscode"
	owner := "microsoft"
	repoName := "vscode"

	// Set up expectations
	mockConfig.EXPECT().Load().Return(config, nil)
	mockConfig.EXPECT().Save(gomock.Any()).DoAndReturn(func(c *services.Config) error {
		// Verify the repo was added
		if len(c.Repos) != 1 {
			t.Errorf("Expected 1 repo, got %d", len(c.Repos))
		}
		if c.Repos[0].Repo != repo_str {
			t.Errorf("Expected '%s', got %s", repo_str, c.Repos[0].Repo)
		}
		return nil
	})
	mockOutput.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()
	mockGitHub.EXPECT().RepoExists(owner, repoName).Return(nil)

	err := cli.handleConfigAdd(repo_str, []string{"stars", "issues"})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestHandleConfigAdd_NotExistingRepo(t *testing.T) {
	ctrl := gomock.NewController(t)

	// testing state
	repo_str := "notexisting/repo"
	owner := "notexisting"
	repoName := "repo"

	mockConfig := mock_services.NewMockConfigService(ctrl)
	mockCache := mock_services.NewMockCacheService(ctrl)
	mockGitHub := mock_services.NewMockGitHubService(ctrl)
	mockOutput := mock_services.NewMockOutput(ctrl)

	cli := NewCLI(mockConfig, mockCache, mockGitHub, mockOutput)

	config := &services.Config{Repos: []services.RepoConfig{}}
	mockConfig.EXPECT().Load().Return(config, nil)
	mockGitHub.EXPECT().RepoExists(owner, repoName).Return(errors.New("not found"))

	err := cli.handleConfigAdd(repo_str, []string{"invalid_event"})

	if err == nil {
		t.Error("Expected error for not existing repo, got nil")
	}
}

func TestHandleConfigAdd_InvalidEvents(t *testing.T) {
	ctrl := gomock.NewController(t)

	// testing state
	repo_str := "microsoft/vscode"
	owner := "microsoft"
	repoName := "vscode"

	mockConfig := mock_services.NewMockConfigService(ctrl)
	mockCache := mock_services.NewMockCacheService(ctrl)
	mockGitHub := mock_services.NewMockGitHubService(ctrl)
	mockOutput := mock_services.NewMockOutput(ctrl)

	cli := NewCLI(mockConfig, mockCache, mockGitHub, mockOutput)

	config := &services.Config{Repos: []services.RepoConfig{}}
	mockConfig.EXPECT().Load().Return(config, nil)
	mockGitHub.EXPECT().RepoExists(owner, repoName).Return(nil)

	err := cli.handleConfigAdd(repo_str, []string{"invalid_event"})

	if err == nil {
		t.Error("Expected error for invalid events, got nil")
	}
}
