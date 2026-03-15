package services

import (
	"fmt"
	"os"
	"path/filepath"

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
