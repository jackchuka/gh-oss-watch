package services

import (
	"context"
	"time"
)

type GitHubServiceImpl struct {
	baseService *GitHubBaseService
	timeout     time.Duration
}

func NewGitHubService() (GitHubService, error) {
	client, err := NewGitHubAPIClient()
	if err != nil {
		return nil, err
	}

	baseService := NewGitHubBaseService(client)
	return &GitHubServiceImpl{
		baseService: baseService,
		timeout:     30 * time.Second,
	}, nil
}

func (g *GitHubServiceImpl) GetRepoStats(owner, repo string) (*RepoStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	return g.baseService.GetRepoStats(ctx, owner, repo)
}

func (g *GitHubServiceImpl) SetMaxConcurrent(maxConcurrent int) {
	// No-op for sequential service
}

func (g *GitHubServiceImpl) SetTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	g.timeout = timeout
}

// RepoExists() is a delegation of RepoExists from baseService
func (g *GitHubServiceImpl) RepoExists(owner, repo string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	return g.baseService.RepoExists(ctx, owner, repo)
}
