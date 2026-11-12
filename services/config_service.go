package services

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigServiceImpl struct{}

func NewConfigService() ConfigService {
	return &ConfigServiceImpl{}
}

func (c *ConfigServiceImpl) Load() (*Config, error) {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Repos: []RepoConfig{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *ConfigServiceImpl) Save(config *Config) error {
	configDir, err := c.getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath, err := c.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *ConfigServiceImpl) GetConfigPath() (string, error) {
	configDir, err := c.getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

func (c *ConfigServiceImpl) getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".gh-oss-watch"), nil
}

func validateEvents(events []string) error {
	validEvents := map[string]bool{
		"stars":         true,
		"issues":        true,
		"pull_requests": true,
		"forks":         true,
		"releases":      true,
	}

	for _, event := range events {
		if !validEvents[event] {
			return fmt.Errorf("invalid event type: %s", event)
		}
	}
	return nil
}

func (c *Config) AddRepo(repo string, events []string, language string) error {
	if err := validateEvents(events); err != nil {
		return err
	}
	for i, r := range c.Repos {
		if r.Repo == repo {
			c.Repos[i].Events = events
			if language != "" {
				c.Repos[i].Language = language
			}
			return nil
		}
	}
	c.Repos = append(c.Repos, RepoConfig{
		Repo:     repo,
		Events:   events,
		Language: language,
	})
	return nil
}

func (c *Config) GetRepo(repo string) *RepoConfig {
	for _, r := range c.Repos {
		if r.Repo == repo {
			return &r
		}
	}
	return nil
}

func (c *Config) RemoveRepo(repo string) error {
	for i, r := range c.Repos {
		if r.Repo == repo {
			c.Repos = append(c.Repos[:i], c.Repos[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("repository %s not found in config", repo)
}

func NormalizeTags(tags []string) ([]string, error) {
	normalized := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || strings.ContainsAny(t, ", \t") {
			return nil, fmt.Errorf("invalid tag %q: must be non-empty without spaces or commas", t)
		}
		if !slices.Contains(normalized, t) {
			normalized = append(normalized, t)
		}
	}
	return normalized, nil
}

func (c *Config) AddTags(repo string, tags []string) error {
	for i, r := range c.Repos {
		if r.Repo == repo {
			for _, t := range tags {
				if !slices.Contains(r.Tags, t) {
					c.Repos[i].Tags = append(c.Repos[i].Tags, t)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("repository %s not found in config", repo)
}

func (c *Config) RemoveTags(repo string, tags []string) error {
	for i, r := range c.Repos {
		if r.Repo == repo {
			c.Repos[i].Tags = slices.DeleteFunc(r.Tags, func(t string) bool {
				return slices.Contains(tags, t)
			})
			if len(c.Repos[i].Tags) == 0 {
				c.Repos[i].Tags = nil
			}
			return nil
		}
	}
	return fmt.Errorf("repository %s not found in config", repo)
}

// FilterByTags returns the repos carrying any of the given tags, or all repos
// when no tags are given. A tag no repo carries is an error, so a typo is not
// mistaken for "nothing to report".
func (c *Config) FilterByTags(tags []string) ([]RepoConfig, error) {
	if len(tags) == 0 {
		return c.Repos, nil
	}
	tags, err := NormalizeTags(tags)
	if err != nil {
		return nil, err
	}

	known := c.AllTags()
	for _, t := range tags {
		if !slices.Contains(known, t) {
			if len(known) == 0 {
				return nil, fmt.Errorf("unknown tag %q: no repos are tagged yet", t)
			}
			return nil, fmt.Errorf("unknown tag %q (known tags: %s)", t, strings.Join(known, ", "))
		}
	}

	var repos []RepoConfig
	for _, r := range c.Repos {
		if slices.ContainsFunc(r.Tags, func(t string) bool { return slices.Contains(tags, t) }) {
			repos = append(repos, r)
		}
	}
	return repos, nil
}

func (c *Config) AllTags() []string {
	var tags []string
	for _, r := range c.Repos {
		for _, t := range r.Tags {
			if !slices.Contains(tags, t) {
				tags = append(tags, t)
			}
		}
	}
	slices.Sort(tags)
	return tags
}
