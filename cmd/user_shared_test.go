/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildUserSyncRequestFromFlags(t *testing.T) {
	req, err := buildUserSyncRequest(userCommandFlags{
		enterprise: "my-enterprise",
		login:      "octocat",
		email:      "octocat@example.com",
		costCenter: "Engineering",
		role:       "direct_member",
		output:     "json",
		orgs:       []string{"platform", "data"},
		teams:      []string{"platform/core", "platform/developers", "data/analysts"},
	})
	if err != nil {
		t.Fatalf("buildUserSyncRequest() error = %v", err)
	}

	if req.Enterprise != "my-enterprise" {
		t.Fatalf("Enterprise = %q, want %q", req.Enterprise, "my-enterprise")
	}

	if req.Login != "octocat" {
		t.Fatalf("Login = %q, want %q", req.Login, "octocat")
	}

	if len(req.Organizations) != 2 {
		t.Fatalf("len(Organizations) = %d, want %d", len(req.Organizations), 2)
	}

	if req.Organizations[0].Name != "platform" {
		t.Fatalf("Organizations[0].Name = %q, want %q", req.Organizations[0].Name, "platform")
	}

	if len(req.Organizations[0].Teams) != 2 {
		t.Fatalf("len(Organizations[0].Teams) = %d, want %d", len(req.Organizations[0].Teams), 2)
	}

	if req.Organizations[1].Name != "data" {
		t.Fatalf("Organizations[1].Name = %q, want %q", req.Organizations[1].Name, "data")
	}

	if len(req.Organizations[1].Teams) != 1 || req.Organizations[1].Teams[0] != "analysts" {
		t.Fatalf("Organizations[1].Teams = %#v, want []string{\"analysts\"}", req.Organizations[1].Teams)
	}
}

func TestBuildUserSyncRequestAllowsEmailOnly(t *testing.T) {
	req, err := buildUserSyncRequest(userCommandFlags{
		enterprise: "my-enterprise",
		email:      "octocat@example.com",
		output:     "json",
		orgs:       []string{"platform"},
		teams:      []string{"platform/core"},
	})
	if err != nil {
		t.Fatalf("buildUserSyncRequest() error = %v", err)
	}

	if req.Login != "" {
		t.Fatalf("Login = %q, want empty string", req.Login)
	}

	if req.Email != "octocat@example.com" {
		t.Fatalf("Email = %q, want %q", req.Email, "octocat@example.com")
	}
}

func TestBuildUserSyncRequestRejectsMixedFileAndDirectInput(t *testing.T) {
	_, err := buildUserSyncRequest(userCommandFlags{
		file:  "request.json",
		login: "octocat",
	})
	if err == nil {
		t.Fatal("buildUserSyncRequest() error = nil, want error")
	}
}

func TestBuildUserSyncRequestFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "request.json")
	content := `{
  "enterprise": "my-enterprise",
  "login": "octocat",
  "organizations": [
    {"name": "platform", "teams": ["core"]}
  ]
}`

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	req, err := buildUserSyncRequest(userCommandFlags{
		file:   path,
		output: "text",
	})
	if err != nil {
		t.Fatalf("buildUserSyncRequest() error = %v", err)
	}

	if req.Enterprise != "my-enterprise" {
		t.Fatalf("Enterprise = %q, want %q", req.Enterprise, "my-enterprise")
	}

	if req.Login != "octocat" {
		t.Fatalf("Login = %q, want %q", req.Login, "octocat")
	}

	if req.CostCenter != "" {
		t.Fatalf("CostCenter = %q, want empty string", req.CostCenter)
	}
}
