/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import "testing"

func TestParseRepositoryReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOrg  string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "valid repo reference",
			input:    "source-org/source-repo",
			wantOrg:  "source-org",
			wantRepo: "source-repo",
		},
		{
			name:    "missing repo segment",
			input:   "source-org",
			wantErr: true,
		},
		{
			name:    "too many segments",
			input:   "source-org/source-repo/extra",
			wantErr: true,
		},
		{
			name:    "empty segment",
			input:   "source-org/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, gotRepo, err := parseRepositoryReference(tt.input, "--Source")
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRepositoryReference() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if gotOrg != tt.wantOrg {
				t.Fatalf("parseRepositoryReference() org = %q, want %q", gotOrg, tt.wantOrg)
			}

			if gotRepo != tt.wantRepo {
				t.Fatalf("parseRepositoryReference() repo = %q, want %q", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestParseOrganizationReference(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOrg string
		wantErr bool
	}{
		{
			name:    "valid org reference",
			input:   "source-org",
			wantOrg: "source-org",
		},
		{
			name:    "repo reference rejected in all mode",
			input:   "source-org/source-repo",
			wantErr: true,
		},
		{
			name:    "blank input",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOrg, err := parseOrganizationReference(tt.input, "--Source")
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOrganizationReference() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if gotOrg != tt.wantOrg {
				t.Fatalf("parseOrganizationReference() org = %q, want %q", gotOrg, tt.wantOrg)
			}
		})
	}
}
