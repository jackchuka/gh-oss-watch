package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
			return c.handleAPIError(err, "")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode >= 400 {
			return c.handleHTTPError(resp.StatusCode, "", nil)
		}

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
		if ghErr, ok := err.(*GitHubError); ok {
			ghErr.Repo = fmt.Sprintf("%s/%s", owner, repo)
			return nil, ghErr
		}
		return nil, NewAPIError("failed to fetch repository data", 0, fmt.Sprintf("%s/%s", owner, repo), err)
	}

	return &repoData, nil
}

func (c *GitHubAPIClientImpl) CheckRepoExists(ctx context.Context, owner, repo string) (bool, error) {
	repoPath := fmt.Sprintf("repos/%s/%s", owner, repo)
	exists := false

	err := WithRetry(ctx, c.retryConfig, func() error {
		resp, err := c.client.RequestWithContext(ctx, "GET", repoPath, nil)
		if err != nil {
			return c.handleAPIError(err, "")
		}
		defer func() {
			// handling error while closing response body
			if err := resp.Body.Close(); err != nil {
				log.Printf("error closing body of response: %v", err)
			}
		}()
		
		
		

		switch resp.StatusCode {
		case 200:
			exists = true
		case 404:
		default:
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *GitHubAPIClientImpl) GetPullRequests(ctx context.Context, owner, repo string) ([]PullRequestAPIData, error) {
	prPath := fmt.Sprintf("repos/%s/%s/pulls?state=open", owner, repo)
	var prs []PullRequestAPIData

	err := c.Get(ctx, prPath, &prs)
	if err != nil {
		if ghErr, ok := err.(*GitHubError); ok {
			ghErr.Repo = fmt.Sprintf("%s/%s", owner, repo)
			return nil, ghErr
		}
		return nil, NewAPIError("failed to fetch pull requests", 0, fmt.Sprintf("%s/%s", owner, repo), err)
	}

	return prs, nil
}

func (c *GitHubAPIClientImpl) handleHTTPError(statusCode int, repo string, err error) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return NewAPIError("authentication failed", statusCode, repo, err)
	case http.StatusForbidden:
		return NewAPIError("access forbidden", statusCode, repo, err)
	case http.StatusNotFound:
		return NewAPIError("resource not found", statusCode, repo, err)
	case http.StatusTooManyRequests:
		return NewAPIError("rate limit exceeded", statusCode, repo, err)
	case http.StatusInternalServerError:
		return NewAPIError("GitHub server error", statusCode, repo, err)
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return NewAPIError("GitHub service unavailable", statusCode, repo, err)
	default:
		return NewAPIError(fmt.Sprintf("HTTP %d error", statusCode), statusCode, repo, err)
	}
}

func (c *GitHubAPIClientImpl) handleAPIError(err error, repo string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline") {
		return NewTimeoutError("request timeout", repo, err)
	}

	if strings.Contains(errStr, "network") || strings.Contains(errStr, "connection") {
		return NewNetworkError("network error", repo, err)
	}

	return fmt.Errorf("GitHub API request failed: %w", err)
}
