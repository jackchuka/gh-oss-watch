package cmd

import (
	"testing"

	"github.com/jackchuka/gh-oss-watch/services"
)

func cfg(repos ...string) *services.Config {
	c := &services.Config{}
	for _, r := range repos {
		c.Repos = append(c.Repos, services.RepoConfig{Repo: r})
	}
	return c
}

func TestValidateSecurityFlags(t *testing.T) {
	config := cfg("jackchuka/timestack")

	if err := validateSecurityFlags(config, "high", "jackchuka/timestack"); err != nil {
		t.Errorf("valid input errored: %v", err)
	}
	if err := validateSecurityFlags(config, "", ""); err != nil {
		t.Errorf("empty input errored: %v", err)
	}
	if err := validateSecurityFlags(config, "bogus", ""); err == nil {
		t.Error("expected error for invalid severity")
	}
	if err := validateSecurityFlags(config, "", "jackchuka/not-watched"); err == nil {
		t.Error("expected error for un-watched repo")
	}
}

func TestSelectSecurityRepos(t *testing.T) {
	config := cfg("o/a", "o/b")
	if got := selectSecurityRepos(config, ""); len(got) != 2 {
		t.Errorf("got %v, want 2 repos", got)
	}
	if got := selectSecurityRepos(config, "o/b"); len(got) != 1 || got[0] != "o/b" {
		t.Errorf("got %v, want [o/b]", got)
	}
}
