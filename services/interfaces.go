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
	GetLatestRelease(ctx context.Context, owner, repo string) (*ReleaseAPIData, error)
	CompareCommits(ctx context.Context, owner, repo, base, head string) (*CommitsComparisonAPIData, error)
	GetStargazers(ctx context.Context, owner, repo string) ([]UserAPIData, error)
}

type GitHubService interface {
	GetRepoStats(owner, repo string) (*RepoStats, error)
	SetMaxConcurrent(maxConcurrent int)
	GetRepoInfo(owner, repo string) (*RepoAPIData, error)
	SetTimeout(timeout time.Duration)
}

type BatchGitHubService interface {
	GitHubService
	GetRepoStatsBatch(repos []string) ([]*RepoStats, []error)
}

type StargazerBatchService interface {
	GetStargazersBatch(repos []string) ([][]UserAPIData, []error)
}

type Output interface {
	Printf(format string, args ...any)
	Println(args ...any)
}

type Config struct {
	Repos []RepoConfig `yaml:"repos"`
}

type RepoConfig struct {
	Repo     string   `yaml:"repo" json:"repo"`
	Events   []string `yaml:"events" json:"events"`
	Language string   `yaml:"language,omitempty" json:"language"`
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
	LastReleaseTag string    `yaml:"last_release_tag,omitempty"`
}

type RepoStats struct {
	Name            string
	Owner           string
	Stars           int
	Issues          int
	PullRequests    int
	Forks           int
	UpdatedAt       time.Time
	LatestRelease   string
	ReleaseDate     time.Time
	UnreleasedCount int
	DefaultBranch   string
	Language        string
}

type EventSummary struct {
	Repo            string `json:"repo"`
	NewStars        int    `json:"newStars"`
	NewIssues       int    `json:"newIssues"`
	NewPRs          int    `json:"newPRs"`
	NewForks        int    `json:"newForks"`
	HasChanges      bool   `json:"hasChanges"`
	NewRelease      bool   `json:"newRelease"`
	UnreleasedCount int    `json:"unreleasedCount"`
	ReleaseTag      string `json:"releaseTag"`
}

type StatusEntry struct {
	EventSummary
	Events      []string `json:"-"`
	TotalStars  int      `json:"totalStars"`
	TotalIssues int      `json:"totalIssues"`
	TotalPRs    int      `json:"totalPRs"`
	TotalForks  int      `json:"totalForks"`
}

type DashboardEntry struct {
	Repo            string    `json:"repo"`
	Stars           int       `json:"stars"`
	Issues          int       `json:"issues"`
	PullRequests    int       `json:"pullRequests"`
	Forks           int       `json:"forks"`
	UpdatedAt       time.Time `json:"updatedAt"`
	LatestRelease   string    `json:"latestRelease,omitempty"`
	UnreleasedCount int       `json:"unreleasedCount,omitempty"`
	Watching        []string  `json:"watching"`
}

type DashboardTotals struct {
	Stars       int `json:"stars"`
	Issues      int `json:"issues"`
	PRs         int `json:"pullRequests"`
	Forks       int `json:"forks"`
	NeedRelease int `json:"needRelease"`
}

type DashboardResult struct {
	Repos  []DashboardEntry `json:"repos"`
	Totals DashboardTotals  `json:"totals"`
}

type ReleaseInfo struct {
	Repo            string    `json:"repo"`
	LatestRelease   string    `json:"latestRelease"`
	ReleaseDate     time.Time `json:"releaseDate"`
	ReleaseAge      string    `json:"releaseAge"`
	UnreleasedCount int       `json:"unreleasedCount"`
	Status          string    `json:"status"`
}

type FanEntry struct {
	Login string   `json:"login"`
	Count int      `json:"count"`
	Repos []string `json:"repos"`
}

type FansResult struct {
	Fans       []FanEntry `json:"fans"`
	TotalFans  int        `json:"totalFans"`
	TotalStars int        `json:"totalStars"`
}

type ListResult struct {
	Repos []RepoConfig `json:"repos"`
	Total int          `json:"total"`
}

type Formatter interface {
	RenderStatus(entries []StatusEntry) error
	RenderDashboard(result DashboardResult) error
	RenderReleases(releases []ReleaseInfo) error
	RenderFans(result FansResult) error
	RenderList(result ListResult) error
}
