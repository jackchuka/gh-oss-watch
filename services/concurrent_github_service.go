package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ConcurrentGitHubService struct {
	baseService *GitHubBaseService
	maxWorkers  int
	timeout     time.Duration
}

type RepoJob struct {
	Owner string
	Repo  string
	Index int
}

type RepoResult struct {
	Stats *RepoStats
	Index int
	Error error
}

func NewConcurrentGitHubService() (BatchGitHubService, error) {
	client, err := NewGitHubAPIClient()
	if err != nil {
		return nil, err
	}

	baseService := NewGitHubBaseService(client)
	return &ConcurrentGitHubService{
		baseService: baseService,
		maxWorkers:  10,
		timeout:     30 * time.Second,
	}, nil
}

func (g *ConcurrentGitHubService) GetRepoInfo(owner, repo string) (*RepoAPIData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	return g.baseService.GetRepoInfo(ctx, owner, repo)
}

func (c *ConcurrentGitHubService) GetRepoStats(owner, repo string) (*RepoStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.baseService.GetRepoStats(ctx, owner, repo)
}

func (c *ConcurrentGitHubService) GetRepoStatsBatch(repos []string) ([]*RepoStats, []error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	jobs := make(chan RepoJob, len(repos))
	results := make(chan RepoResult, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg, jobs, results)
	}

	go func() {
		defer close(jobs)
		for i, repoStr := range repos {
			owner, repo, err := ParseRepoString(repoStr)
			if err != nil {
				results <- RepoResult{
					Stats: nil,
					Index: i,
					Error: fmt.Errorf("invalid repo format %s: %w", repoStr, err),
				}
				continue
			}

			select {
			case jobs <- RepoJob{Owner: owner, Repo: repo, Index: i}:
			case <-ctx.Done():
				results <- RepoResult{
					Stats: nil,
					Index: i,
					Error: ctx.Err(),
				}
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	stats := make([]*RepoStats, len(repos))
	errors := make([]error, len(repos))

	for result := range results {
		if result.Index < len(repos) {
			stats[result.Index] = result.Stats
			errors[result.Index] = result.Error
		}
	}

	return stats, errors
}

func (c *ConcurrentGitHubService) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan RepoJob, results chan<- RepoResult) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}

			stats, err := c.baseService.GetRepoStats(ctx, job.Owner, job.Repo)
			results <- RepoResult{
				Stats: stats,
				Index: job.Index,
				Error: err,
			}

		case <-ctx.Done():
			return
		}
	}
}

type StargazerJob struct {
	Owner string
	Repo  string
	Index int
}

type StargazerResult struct {
	Users []UserAPIData
	Repo  string
	Index int
	Error error
}

func (c *ConcurrentGitHubService) GetStargazersBatch(repos []string) ([][]UserAPIData, []error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	jobs := make(chan StargazerJob, len(repos))
	results := make(chan StargazerResult, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.stargazerWorker(ctx, &wg, jobs, results)
	}

	go func() {
		defer close(jobs)
		for i, repoStr := range repos {
			owner, repo, err := ParseRepoString(repoStr)
			if err != nil {
				results <- StargazerResult{
					Repo:  repoStr,
					Index: i,
					Error: fmt.Errorf("invalid repo format %s: %w", repoStr, err),
				}
				continue
			}

			select {
			case jobs <- StargazerJob{Owner: owner, Repo: repo, Index: i}:
			case <-ctx.Done():
				results <- StargazerResult{
					Repo:  repoStr,
					Index: i,
					Error: ctx.Err(),
				}
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	users := make([][]UserAPIData, len(repos))
	errors := make([]error, len(repos))

	for result := range results {
		if result.Index < len(repos) {
			users[result.Index] = result.Users
			errors[result.Index] = result.Error
		}
	}

	return users, errors
}

func (c *ConcurrentGitHubService) stargazerWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan StargazerJob, results chan<- StargazerResult) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}

			stargazers, err := c.baseService.client.GetStargazers(ctx, job.Owner, job.Repo)
			results <- StargazerResult{
				Users: stargazers,
				Repo:  fmt.Sprintf("%s/%s", job.Owner, job.Repo),
				Index: job.Index,
				Error: err,
			}

		case <-ctx.Done():
			return
		}
	}
}

type AlertJob struct {
	Owner string
	Repo  string
	Index int
}

type AlertResult struct {
	Alerts []SecurityAlert
	Index  int
	Error  error
}

func (c *ConcurrentGitHubService) GetDependabotAlertsBatch(repos []string) ([][]SecurityAlert, []error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	jobs := make(chan AlertJob, len(repos))
	results := make(chan AlertResult, len(repos))

	var wg sync.WaitGroup
	for i := 0; i < c.maxWorkers; i++ {
		wg.Add(1)
		go c.alertWorker(ctx, &wg, jobs, results)
	}

	go func() {
		defer close(jobs)
		for i, repoStr := range repos {
			owner, repo, err := ParseRepoString(repoStr)
			if err != nil {
				results <- AlertResult{
					Index: i,
					Error: fmt.Errorf("invalid repo format %s: %w", repoStr, err),
				}
				continue
			}
			select {
			case jobs <- AlertJob{Owner: owner, Repo: repo, Index: i}:
			case <-ctx.Done():
				results <- AlertResult{
					Index: i,
					Error: ctx.Err(),
				}
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	alerts := make([][]SecurityAlert, len(repos))
	errs := make([]error, len(repos))
	for r := range results {
		if r.Index < len(repos) {
			alerts[r.Index] = r.Alerts
			errs[r.Index] = r.Error
		}
	}

	return alerts, errs
}

func (c *ConcurrentGitHubService) alertWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan AlertJob, results chan<- AlertResult) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			raw, err := c.baseService.client.GetDependabotAlerts(ctx, job.Owner, job.Repo)
			converted := make([]SecurityAlert, 0, len(raw))
			for _, a := range raw {
				converted = append(converted, toSecurityAlert(a))
			}
			results <- AlertResult{
				Alerts: converted,
				Index:  job.Index,
				Error:  err,
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *ConcurrentGitHubService) SetMaxConcurrent(maxConcurrent int) {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	c.maxWorkers = maxConcurrent
}

func (c *ConcurrentGitHubService) SetTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	c.timeout = timeout
}
