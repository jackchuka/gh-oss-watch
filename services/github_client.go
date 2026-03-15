package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	DefaultBranch   string    `json:"default_branch"`
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

type ReleaseAPIData struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

type CommitsComparisonAPIData struct {
	AheadBy      int `json:"ahead_by"`
	TotalCommits int `json:"total_commits"`
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

func (c *GitHubAPIClientImpl) GetLatestRelease(ctx context.Context, owner, repo string) (*ReleaseAPIData, error) {
	releasePath := fmt.Sprintf("repos/%s/%s/releases/latest", owner, repo)
	var release ReleaseAPIData

	err := c.Get(ctx, releasePath, &release)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
			return nil, nil // no releases exist
		}
		var ghErr *GitHubError
		if errors.As(err, &ghErr) && ghErr.Type == ErrorTypeNotFound {
			return nil, nil // no releases exist
		}
		return nil, err
	}

	return &release, nil
}

func (c *GitHubAPIClientImpl) CompareCommits(ctx context.Context, owner, repo, base, head string) (*CommitsComparisonAPIData, error) {
	comparePath := fmt.Sprintf("repos/%s/%s/compare/%s...%s", owner, repo, base, head)
	var comparison CommitsComparisonAPIData

	err := c.Get(ctx, comparePath, &comparison)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			return nil, NewAPIError("failed to compare commits", httpErr.StatusCode, fmt.Sprintf("%s/%s", owner, repo), err)
		}
		return nil, err
	}

	return &comparison, nil
}

func (c *GitHubAPIClientImpl) getPaginated(ctx context.Context, path string) ([]UserAPIData, error) {
	var all []UserAPIData
	currentPath := path

	for currentPath != "" {
		var page []UserAPIData
		nextPath, err := c.getPage(ctx, currentPath, &page)
		if err != nil {
			return all, err
		}
		all = append(all, page...)
		currentPath = nextPath
	}

	return all, nil
}

func (c *GitHubAPIClientImpl) getPage(ctx context.Context, path string, result *[]UserAPIData) (string, error) {
	var nextPath string

	err := WithRetry(ctx, c.retryConfig, func() error {
		resp, err := c.client.RequestWithContext(ctx, "GET", path, nil)
		if err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(result); err != nil {
			return NewAPIError("failed to decode JSON response", resp.StatusCode, "", err)
		}

		nextPath = parseNextLink(resp.Header.Get("Link"))
		return nil
	})

	return nextPath, err
}

func parseNextLink(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			start := strings.Index(part, "<")
			end := strings.Index(part, ">")
			if start >= 0 && end > start {
				link := part[start+1 : end]
				if idx := strings.Index(link, "/repos/"); idx >= 0 {
					return link[idx+1:]
				}
			}
		}
	}
	return ""
}

func (c *GitHubAPIClientImpl) GetStargazers(ctx context.Context, owner, repo string) ([]UserAPIData, error) {
	path := fmt.Sprintf("repos/%s/%s/stargazers?per_page=100", owner, repo)

	users, err := c.getPaginated(ctx, path)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			return nil, NewAPIError("failed to fetch stargazers", httpErr.StatusCode, fmt.Sprintf("%s/%s", owner, repo), err)
		}
		return nil, err
	}

	return users, nil
}
