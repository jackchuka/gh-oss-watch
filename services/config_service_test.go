package services

import (
	"slices"
	"testing"
)

func TestFilterByTags(t *testing.T) {
	config := &Config{Repos: []RepoConfig{
		{Repo: "acme/api", Tags: []string{"work"}},
		{Repo: "me/cli", Tags: []string{"personal", "go"}},
		{Repo: "me/notes"},
	}}

	tests := []struct {
		name    string
		tags    []string
		want    []string
		wantErr bool
	}{
		{name: "no tags returns all repos", want: []string{"acme/api", "me/cli", "me/notes"}},
		{name: "single tag", tags: []string{"work"}, want: []string{"acme/api"}},
		{name: "any of several tags, case-insensitive", tags: []string{"WORK", "go"}, want: []string{"acme/api", "me/cli"}},
		{name: "unknown tag errors", tags: []string{"wrok"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repos, err := config.FilterByTags(tt.tags)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			var got []string
			for _, r := range repos {
				got = append(got, r.Repo)
			}
			if !tt.wantErr && !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	got, err := NormalizeTags([]string{" Work ", "work", "OSS"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"work", "oss"}) {
		t.Errorf("got %v", got)
	}
	for _, bad := range []string{"", "a b", "a,b"} {
		if _, err := NormalizeTags([]string{bad}); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}
