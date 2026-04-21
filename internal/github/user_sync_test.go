package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gogithub "github.com/google/go-github/v84/github"
)

func TestNormalizeUserSyncRequest(t *testing.T) {
	req, err := NormalizeUserSyncRequest(UserSyncRequest{
		Enterprise: "  enterprise  ",
		Login:      "  octocat  ",
		Organizations: []UserOrganizationAssignment{
			{Name: " platform ", Teams: []string{" core ", "developers"}},
		},
	})
	if err != nil {
		t.Fatalf("NormalizeUserSyncRequest() error = %v", err)
	}

	if req.Role != "direct_member" {
		t.Fatalf("Role = %q, want %q", req.Role, "direct_member")
	}

	if req.Enterprise != "enterprise" {
		t.Fatalf("Enterprise = %q, want %q", req.Enterprise, "enterprise")
	}

	if req.Login != "octocat" {
		t.Fatalf("Login = %q, want %q", req.Login, "octocat")
	}

	if req.Organizations[0].Name != "platform" {
		t.Fatalf("Organizations[0].Name = %q, want %q", req.Organizations[0].Name, "platform")
	}

	if req.Organizations[0].Teams[0] != "core" {
		t.Fatalf("Organizations[0].Teams[0] = %q, want %q", req.Organizations[0].Teams[0], "core")
	}
}

func TestNormalizeUserSyncRequestAllowsEmailOnly(t *testing.T) {
	req, err := NormalizeUserSyncRequest(UserSyncRequest{
		Enterprise: "enterprise",
		Email:      "octocat@example.com",
		Organizations: []UserOrganizationAssignment{
			{Name: "platform", Teams: []string{"core"}},
		},
	})
	if err != nil {
		t.Fatalf("NormalizeUserSyncRequest() error = %v", err)
	}

	if req.Login != "" {
		t.Fatalf("Login = %q, want empty string", req.Login)
	}

	if req.Email != "octocat@example.com" {
		t.Fatalf("Email = %q, want %q", req.Email, "octocat@example.com")
	}
}

func TestAddUserCreatesInvitation(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/users/octocat", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"id":    101,
			"login": "octocat",
		})
	})

	mux.HandleFunc("/orgs/platform/teams/core", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"id":   11,
			"slug": "core",
		})
	})

	mux.HandleFunc("/orgs/platform/memberships/octocat", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	mux.HandleFunc("/orgs/platform/invitations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, []map[string]any{})
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Decode(invitation body) error = %v", err)
			}

			if got := body["role"]; got != "direct_member" {
				t.Fatalf("role = %v, want %q", got, "direct_member")
			}

			teamIDs, ok := body["team_ids"].([]any)
			if !ok || len(teamIDs) != 1 {
				t.Fatalf("team_ids = %#v, want one team ID", body["team_ids"])
			}

			writeJSON(t, w, map[string]any{
				"id":    1,
				"login": "octocat",
			})
		default:
			t.Fatalf("unexpected method %s for invitations", r.Method)
		}
	})

	client := newTestGitHubClient(t, server)

	result, err := AddUser(client, UserSyncRequest{
		Enterprise: "my-enterprise",
		Login:      "octocat",
		Organizations: []UserOrganizationAssignment{
			{Name: "platform", Teams: []string{"core"}},
		},
	}, UserSyncOptions{})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	if result.Organizations[0].Membership != "invite-created" {
		t.Fatalf("Membership = %q, want %q", result.Organizations[0].Membership, "invite-created")
	}

	if result.Organizations[0].Teams[0].Action != "queued-with-invite" {
		t.Fatalf("Team action = %q, want %q", result.Organizations[0].Teams[0].Action, "queued-with-invite")
	}

	if result.CostCenter != nil {
		t.Fatalf("CostCenter = %#v, want nil", result.CostCenter)
	}
}

func TestAddUserAllowsEmailOnlyInvitation(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/orgs/platform/teams/core", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"id":   11,
			"slug": "core",
		})
	})

	mux.HandleFunc("/orgs/platform/invitations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(t, w, []map[string]any{})
		case http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Decode(invitation body) error = %v", err)
			}

			if got := body["email"]; got != "octocat@example.com" {
				t.Fatalf("email = %v, want %q", got, "octocat@example.com")
			}

			if _, exists := body["invitee_id"]; exists {
				t.Fatalf("invitee_id should be omitted, got %#v", body["invitee_id"])
			}

			writeJSON(t, w, map[string]any{
				"id":    1,
				"email": "octocat@example.com",
			})
		default:
			t.Fatalf("unexpected method %s for invitations", r.Method)
		}
	})

	client := newTestGitHubClient(t, server)

	result, err := AddUser(client, UserSyncRequest{
		Enterprise: "my-enterprise",
		Email:      "octocat@example.com",
		Organizations: []UserOrganizationAssignment{
			{Name: "platform", Teams: []string{"core"}},
		},
	}, UserSyncOptions{})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	if result.Organizations[0].Membership != "invite-created" {
		t.Fatalf("Membership = %q, want %q", result.Organizations[0].Membership, "invite-created")
	}

	if result.Organizations[0].Teams[0].Action != "queued-with-invite" {
		t.Fatalf("Team action = %q, want %q", result.Organizations[0].Teams[0].Action, "queued-with-invite")
	}

	if result.CostCenter != nil {
		t.Fatalf("CostCenter = %#v, want nil", result.CostCenter)
	}
}

func TestUpdateUserEnsuresTeamAndReassignsCostCenter(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/users/octocat", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"id":    101,
			"login": "octocat",
		})
	})

	mux.HandleFunc("/enterprises/my-enterprise/settings/billing/cost-centers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"costCenters": []map[string]any{
				{
					"id":    "cc-eng",
					"name":  "Engineering",
					"state": "active",
					"resources": []map[string]any{
						{"type": "User", "name": "someone-else"},
					},
				},
				{
					"id":    "cc-legacy",
					"name":  "Legacy",
					"state": "active",
					"resources": []map[string]any{
						{"type": "User", "name": "octocat"},
					},
				},
			},
		})
	})

	mux.HandleFunc("/orgs/platform/teams/core", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"id":   11,
			"slug": "core",
		})
	})

	mux.HandleFunc("/orgs/platform/memberships/octocat", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"state": "active",
			"role":  "member",
		})
	})

	mux.HandleFunc("/orgs/platform/teams/core/memberships/octocat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("team membership method = %s, want PUT", r.Method)
		}

		writeJSON(t, w, map[string]any{
			"state": "active",
			"role":  "member",
		})
	})

	mux.HandleFunc("/enterprises/my-enterprise/settings/billing/cost-centers/cc-eng/resource", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"message": "Resources successfully added to the cost center.",
			"reassigned_resources": []map[string]any{
				{
					"resource_type":        "user",
					"name":                 "octocat",
					"previous_cost_center": "Legacy",
				},
			},
		})
	})

	client := newTestGitHubClient(t, server)

	result, err := UpdateUser(client, UserSyncRequest{
		Enterprise: "my-enterprise",
		Login:      "octocat",
		CostCenter: "Engineering",
		Organizations: []UserOrganizationAssignment{
			{Name: "platform", Teams: []string{"core"}},
		},
	}, UserSyncOptions{})
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	if result.Organizations[0].Membership != "already-active" {
		t.Fatalf("Membership = %q, want %q", result.Organizations[0].Membership, "already-active")
	}

	if result.Organizations[0].Teams[0].Action != "ensured" {
		t.Fatalf("Team action = %q, want %q", result.Organizations[0].Teams[0].Action, "ensured")
	}

	if result.CostCenter.Action != "reassigned" {
		t.Fatalf("CostCenter.Action = %q, want %q", result.CostCenter.Action, "reassigned")
	}

	if result.CostCenter.PreviousCostCenter != "Legacy" {
		t.Fatalf("PreviousCostCenter = %q, want %q", result.CostCenter.PreviousCostCenter, "Legacy")
	}
}

func TestUpdateUserRequiresLogin(t *testing.T) {
	client := gogithub.NewClient(nil)

	_, err := UpdateUser(client, UserSyncRequest{
		Enterprise: "my-enterprise",
		Email:      "octocat@example.com",
		CostCenter: "Engineering",
		Organizations: []UserOrganizationAssignment{
			{Name: "platform"},
		},
	}, UserSyncOptions{})
	if err == nil {
		t.Fatal("UpdateUser() error = nil, want error")
	}
}

func newTestGitHubClient(t *testing.T, server *httptest.Server) *gogithub.Client {
	t.Helper()

	client := gogithub.NewClient(server.Client())
	baseURL := server.URL + "/"

	var err error
	client.BaseURL, err = client.BaseURL.Parse(baseURL)
	if err != nil {
		t.Fatalf("Parse(baseURL) error = %v", err)
	}

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}
