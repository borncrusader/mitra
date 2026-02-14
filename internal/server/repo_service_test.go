package server

import (
	"testing"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantHost  string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "GitHub with https",
			url:       "https://github.com/owner/repo",
			wantHost:  "github.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "GitHub without protocol",
			url:       "github.com/owner/repo",
			wantHost:  "github.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "GitLab with https",
			url:       "https://gitlab.com/owner/repo",
			wantHost:  "gitlab.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Bitbucket",
			url:       "bitbucket.org/owner/repo",
			wantHost:  "bitbucket.org",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "Custom domain",
			url:       "git.company.com/owner/repo",
			wantHost:  "git.company.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "with .git suffix",
			url:       "https://github.com/owner/repo.git",
			wantHost:  "github.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "with trailing slash",
			url:       "github.com/owner/repo/",
			wantHost:  "github.com",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:    "invalid - only owner/repo",
			url:     "owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid - only host/owner",
			url:     "github.com/owner",
			wantErr: true,
		},
		{
			name:    "invalid - too many parts",
			url:     "github.com/owner/repo/extra",
			wantErr: true,
		},
		{
			name:    "invalid - empty owner",
			url:     "github.com//repo",
			wantErr: true,
		},
		{
			name:    "invalid - empty repo",
			url:     "github.com/owner/",
			wantErr: true,
		},
		{
			name:    "invalid - no path",
			url:     "github.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, owner, repo, err := parseGitURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseGitURL() expected error but got none for URL: %s", tt.url)
				}
				return
			}

			if err != nil {
				t.Errorf("parseGitURL() unexpected error: %v", err)
				return
			}

			if host != tt.wantHost {
				t.Errorf("parseGitURL() host = %q, want %q", host, tt.wantHost)
			}

			if owner != tt.wantOwner {
				t.Errorf("parseGitURL() owner = %q, want %q", owner, tt.wantOwner)
			}

			if repo != tt.wantRepo {
				t.Errorf("parseGitURL() repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}
