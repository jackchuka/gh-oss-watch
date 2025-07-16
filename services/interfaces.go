package services

//go:generate go tool mockgen -source=$GOFILE -package=mock_$GOPACKAGE -destination=./mock/mock_$GOFILE

import (
	"context"
	"time"
)

type ConfigService interface {
	Load() (*Config, error)
	Save(config *Config) error
	GetConfigPath() (string, error)
}

type CacheService interface {
	Load() (*CacheData, error)
	Save(cache *CacheData) error
}

type GitHubAPIClient interface {
	Get(ctx context.Context, path string, response any) error
	GetRepoData(ctx context.Context, owner, repo string) (*RepoAPIData, error)
	GetPullRequests(ctx context.Context, owner, repo string) ([]PullRequestAPIData, error)
	CheckRepoExists(ctx context.Context, owner, repo string) (bool, error)
}

type GitHubService interface {
	GetRepoStats(owner, repo string) (*RepoStats, error)
	SetMaxConcurrent(maxConcurrent int)
	SetTimeout(timeout time.Duration)
	RepoExists(owner, repo string) (bool, error)
}

type BatchGitHubService interface {
	GitHubService
	GetRepoStatsBatch(repos []string) ([]*RepoStats, []error)
}

type Output interface {
	Printf(format string, args ...any)
	Println(args ...any)
}

type Config struct {
	Repos []RepoConfig `yaml:"repos"`
}

type RepoConfig struct {
	Repo   string   `yaml:"repo"`
	Events []string `yaml:"events"`
}

type CacheData struct {
	LastCheck time.Time            `yaml:"last_check"`
	Repos     map[string]RepoState `yaml:"repos"`
}

type RepoState struct {
	LastStarCount  int       `yaml:"last_star_count"`
	LastIssueCount int       `yaml:"last_issue_count"`
	LastPRCount    int       `yaml:"last_pr_count"`
	LastForkCount  int       `yaml:"last_fork_count"`
	LastUpdated    time.Time `yaml:"last_updated"`
}

type RepoStats struct {
	Name         string
	Owner        string
	Stars        int
	Issues       int
	PullRequests int
	Forks        int
	UpdatedAt    time.Time
}

type EventSummary struct {
	Repo       string
	NewStars   int
	NewIssues  int
	NewPRs     int
	NewForks   int
	HasChanges bool
}
