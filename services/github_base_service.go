package services

import (
	"context"
	"fmt"
	"strings"
)

// GitHubBaseService provides common GitHub operations for both single and concurrent services
type GitHubBaseService struct {
	client GitHubAPIClient
}

// NewGitHubBaseService creates a new base GitHub service
func NewGitHubBaseService(client GitHubAPIClient) *GitHubBaseService {
	return &GitHubBaseService{
		client: client,
	}
}

// GetRepoStats fetches repository statistics for a single repository
func (g *GitHubBaseService) GetRepoStats(ctx context.Context, owner, repo string) (*RepoStats, error) {
	repoData, err := g.client.GetRepoData(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	prs, err := g.client.GetPullRequests(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &RepoStats{
		Name:         repoData.Name,
		Owner:        repoData.Owner.Login,
		Stars:        repoData.StargazersCount,
		Issues:       repoData.OpenIssuesCount,
		PullRequests: len(prs),
		Forks:        repoData.ForksCount,
		UpdatedAt:    repoData.UpdatedAt,
	}, nil
}

// RepoExists() checkes repo existance by fetching it
func (g *GitHubBaseService) RepoExists(ctx context.Context, owner, repo string) (bool, error) {
	return g.client.CheckRepoExists(ctx, owner, repo)
}

// ParseRepoString parses a repository string in the format "owner/repo"
func ParseRepoString(repoStr string) (owner, repo string, err error) {
	parts := strings.Split(repoStr, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", NewValidationError(
			fmt.Sprintf("invalid repo format: %s (expected owner/repo)", repoStr),
			repoStr,
			nil,
		)
	}
	return parts[0], parts[1], nil
}

// CalculateEventSummary compares current stats with previous state to determine changes
func CalculateEventSummary(repoStr string, current *RepoStats, previous RepoState) EventSummary {
	summary := EventSummary{
		Repo: repoStr,
	}

	if current.Stars > previous.LastStarCount {
		summary.NewStars = current.Stars - previous.LastStarCount
		summary.HasChanges = true
	}

	if current.Issues > previous.LastIssueCount {
		summary.NewIssues = current.Issues - previous.LastIssueCount
		summary.HasChanges = true
	}

	if current.PullRequests > previous.LastPRCount {
		summary.NewPRs = current.PullRequests - previous.LastPRCount
		summary.HasChanges = true
	}

	if current.Forks > previous.LastForkCount {
		summary.NewForks = current.Forks - previous.LastForkCount
		summary.HasChanges = true
	}

	return summary
}
