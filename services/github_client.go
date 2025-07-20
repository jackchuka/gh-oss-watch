package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

type RepoAPIData struct {
	Name            string    `json:"name"`
	Owner           OwnerData `json:"owner"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OwnerData struct {
	Login string `json:"login"`
}

type PullRequestAPIData struct {
	ID     int    `json:"id"`
	Number int    `json:"number"`
	State  string `json:"state"`
	Title  string `json:"title"`
}

type UserAPIData struct {
	Login string `json:"login"`
}

type GitHubAPIClientImpl struct {
	client      *api.RESTClient
	retryConfig RetryConfig
}

func NewGitHubAPIClient() (GitHubAPIClient, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return nil, NewConfigError("failed to create GitHub API client", err)
	}

	return &GitHubAPIClientImpl{
		client:      restClient,
		retryConfig: DefaultRetryConfig(),
	}, nil
}

func (c *GitHubAPIClientImpl) Get(ctx context.Context, path string, response any) error {
	return WithRetry(ctx, c.retryConfig, func() error {
		resp, err := c.client.RequestWithContext(ctx, "GET", path, nil)
		if err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(response); err != nil {
			return NewAPIError("failed to decode JSON response", resp.StatusCode, "", err)
		}

		return nil
	})
}

func (c *GitHubAPIClientImpl) GetRepoData(ctx context.Context, owner, repo string) (*RepoAPIData, error) {
	repoPath := fmt.Sprintf("repos/%s/%s", owner, repo)
	var repoData RepoAPIData

	err := c.Get(ctx, repoPath, &repoData)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			return nil, NewAPIError("failed to fetch repository data", httpErr.StatusCode, fmt.Sprintf("%s/%s", owner, repo), err)
		}
		return nil, err
	}

	return &repoData, nil
}

func (c *GitHubAPIClientImpl) GetPullRequests(ctx context.Context, owner, repo string) ([]PullRequestAPIData, error) {
	prPath := fmt.Sprintf("repos/%s/%s/pulls?state=open", owner, repo)
	var prs []PullRequestAPIData

	err := c.Get(ctx, prPath, &prs)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			return nil, NewAPIError("failed to fetch pull requests", httpErr.StatusCode, fmt.Sprintf("%s/%s", owner, repo), err)
		}
		return nil, err
	}

	return prs, nil
}
