package services

import "testing"

func TestParseNextLink(t *testing.T) {
	tests := []struct {
		name       string
		linkHeader string
		want       string
	}{
		{
			name:       "extracts next link from typical GitHub Link header",
			linkHeader: `<https://api.github.com/repos/owner/repo/stargazers?page=2&per_page=100>; rel="next", <https://api.github.com/repos/owner/repo/stargazers?page=5&per_page=100>; rel="last"`,
			want:       "repos/owner/repo/stargazers?page=2&per_page=100",
		},
		{
			name:       "returns empty for last page (no next)",
			linkHeader: `<https://api.github.com/repos/owner/repo/stargazers?page=1&per_page=100>; rel="first", <https://api.github.com/repos/owner/repo/stargazers?page=4&per_page=100>; rel="prev"`,
			want:       "",
		},
		{
			name:       "returns empty for empty header",
			linkHeader: "",
			want:       "",
		},
		{
			name:       "handles single next link",
			linkHeader: `<https://api.github.com/repos/owner/repo/stargazers?page=3&per_page=100>; rel="next"`,
			want:       "repos/owner/repo/stargazers?page=3&per_page=100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNextLink(tt.linkHeader)
			if got != tt.want {
				t.Errorf("parseNextLink() = %q, want %q", got, tt.want)
			}
		})
	}
}
