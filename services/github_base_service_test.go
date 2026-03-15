package services

import "testing"

func TestParseRepoString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "owner/repo",
			input:     "microsoft/vscode",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "full https URL",
			input:     "https://github.com/microsoft/vscode",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "full https URL with trailing slash",
			input:     "https://github.com/microsoft/vscode/",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "http URL",
			input:     "http://github.com/microsoft/vscode",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      ".git suffix",
			input:     "github.com/microsoft/vscode.git",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "go import style",
			input:     "github.com/microsoft/vscode",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "https URL with .git suffix",
			input:     "https://github.com/microsoft/vscode.git",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:      "whitespace around input",
			input:     "  microsoft/vscode  ",
			wantOwner: "microsoft",
			wantRepo:  "vscode",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "just a name with no slash",
			input:   "vscode",
			wantErr: true,
		},
		{
			name:    "too many path segments",
			input:   "https://github.com/microsoft/vscode/tree/main",
			wantErr: true,
		},
		{
			name:    "missing repo",
			input:   "microsoft/",
			wantErr: true,
		},
		{
			name:    "missing owner",
			input:   "/vscode",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoString(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got owner=%q repo=%q", owner, repo)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}
