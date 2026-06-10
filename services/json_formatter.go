package services

import (
	"encoding/json"
	"io"
)

type JSONFormatter struct {
	w io.Writer
}

func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

func (f *JSONFormatter) RenderStatus(entries []StatusEntry) error {
	if entries == nil {
		entries = []StatusEntry{}
	}
	return json.NewEncoder(f.w).Encode(entries)
}

func (f *JSONFormatter) RenderDashboard(result DashboardResult) error {
	return json.NewEncoder(f.w).Encode(result)
}

func (f *JSONFormatter) RenderReleases(releases []ReleaseInfo) error {
	if releases == nil {
		releases = []ReleaseInfo{}
	}
	return json.NewEncoder(f.w).Encode(releases)
}

func (f *JSONFormatter) RenderFans(result FansResult) error {
	if result.Fans == nil {
		result.Fans = []FanEntry{}
	}
	return json.NewEncoder(f.w).Encode(result)
}

func (f *JSONFormatter) RenderList(result ListResult) error {
	if result.Repos == nil {
		result.Repos = []RepoConfig{}
	}
	return json.NewEncoder(f.w).Encode(result)
}

func (f *JSONFormatter) RenderSecurity(result SecurityResult, _ bool) error {
	if result.Repos == nil {
		result.Repos = []SecurityRepoEntry{}
	}
	if result.SkippedRepos == nil {
		result.SkippedRepos = []string{}
	}
	if result.Totals == nil {
		result.Totals = map[string]int{}
	}
	return json.NewEncoder(f.w).Encode(result)
}
