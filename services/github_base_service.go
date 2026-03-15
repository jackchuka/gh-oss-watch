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

	stats := &RepoStats{
		Name:          repoData.Name,
		Owner:         repoData.Owner.Login,
		Stars:         repoData.StargazersCount,
		Issues:        repoData.OpenIssuesCount,
		PullRequests:  len(prs),
		Forks:         repoData.ForksCount,
		UpdatedAt:     repoData.UpdatedAt,
		DefaultBranch: repoData.DefaultBranch,
		Language:      repoData.Language,
	}

	release, err := g.client.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		// Non-fatal: release info is optional
		return stats, nil
	}

	if release != nil {
		stats.LatestRelease = release.TagName
		stats.ReleaseDate = release.PublishedAt

		comparison, err := g.client.CompareCommits(ctx, owner, repo, release.TagName, repoData.DefaultBranch)
		if err == nil {
			stats.UnreleasedCount = comparison.AheadBy
		}
	}

	return stats, nil
}

func (g *GitHubBaseService) GetRepoInfo(ctx context.Context, owner, repo string) (*RepoAPIData, error) {
	return g.client.GetRepoData(ctx, owner, repo)
}

// ParseRepoString parses a repository string into owner and repo.
// Accepts: "owner/repo", "https://github.com/owner/repo", "github.com/owner/repo.git", etc.
func ParseRepoString(repoStr string) (owner, repo string, err error) {
	s := strings.TrimSpace(repoStr)
	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "github.com/")
	s = strings.TrimSuffix(s, "/")

	parts := strings.Split(s, "/")
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

	if current.LatestRelease != "" {
		summary.ReleaseTag = current.LatestRelease
		summary.UnreleasedCount = current.UnreleasedCount
		if current.LatestRelease != previous.LastReleaseTag {
			summary.NewRelease = true
		}
	}

	return summary
}
