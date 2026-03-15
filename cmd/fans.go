package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/jackchuka/gh-oss-watch/services"
)

func aggregateFans(repoUsers map[string][]services.UserAPIData) services.FansResult {
	type fanData struct {
		repos map[string]bool
	}

	fanMap := make(map[string]*fanData)
	totalStars := 0

	for repo, users := range repoUsers {
		for _, user := range users {
			totalStars++
			if _, ok := fanMap[user.Login]; !ok {
				fanMap[user.Login] = &fanData{repos: make(map[string]bool)}
			}
			fanMap[user.Login].repos[repo] = true
		}
	}

	fans := make([]services.FanEntry, 0, len(fanMap))
	for login, data := range fanMap {
		repos := make([]string, 0, len(data.repos))
		for repo := range data.repos {
			repos = append(repos, repo)
		}
		sort.Strings(repos)
		fans = append(fans, services.FanEntry{
			Login: login,
			Count: len(data.repos),
			Repos: repos,
		})
	}

	sort.Slice(fans, func(i, j int) bool {
		if fans[i].Count != fans[j].Count {
			return fans[i].Count > fans[j].Count
		}
		// Within same count, sort by first repo name
		if fans[i].Repos[0] != fans[j].Repos[0] {
			return fans[i].Repos[0] < fans[j].Repos[0]
		}
		return fans[i].Login < fans[j].Login
	})

	return services.FansResult{
		Fans:       fans,
		TotalFans:  len(fans),
		TotalStars: totalStars,
	}
}

func (c *CLI) handleFans(top int) error {
	config, err := c.validateConfig()
	if err != nil {
		return err
	}

	if len(config.Repos) == 0 {
		return nil
	}

	stargazerService, ok := c.githubService.(services.StargazerBatchService)
	if !ok {
		return fmt.Errorf("fans command requires a service that supports batch stargazer fetching")
	}

	repos := make([]string, len(config.Repos))
	for i, rc := range config.Repos {
		repos[i] = rc.Repo
	}

	allUsers, allErrors := stargazerService.GetStargazersBatch(repos)

	repoUsers := make(map[string][]services.UserAPIData)
	for i, rc := range config.Repos {
		if allErrors[i] != nil {
			fmt.Fprintf(os.Stderr, "Error fetching stargazers for %s: %v\n", rc.Repo, allErrors[i])
			continue
		}
		if allUsers[i] != nil {
			repoUsers[rc.Repo] = allUsers[i]
		}
	}

	result := aggregateFans(repoUsers)
	if top > 0 && top < len(result.Fans) {
		result.Fans = result.Fans[:top]
	}
	return c.formatter.RenderFans(result)
}
