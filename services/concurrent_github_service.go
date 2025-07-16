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

// RepoExists() is a delegation of RepoExists from baseService
func (g *ConcurrentGitHubService) RepoExists(owner, repo string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	return g.baseService.RepoExists(ctx, owner, repo)
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
