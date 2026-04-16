package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gogithub "github.com/google/go-github/v84/github"
)

func TestNormalizeVisibilityFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "empty defaults to all",
			input: "",
			want:  "all",
		},
		{
			name:  "all is preserved",
			input: "all",
			want:  "all",
		},
		{
			name:  "public is normalized",
			input: "PUBLIC",
			want:  "public",
		},
		{
			name:  "private trims whitespace",
			input: "  private  ",
			want:  "private",
		},
		{
			name:    "invalid value returns error",
			input:   "internal",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeVisibilityFilter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeVisibilityFilter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Fatalf("normalizeVisibilityFilter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestListReposPaginatesAcrossAllPages(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("type"); got != "all" {
			t.Fatalf("type query = %q, want %q", got, "all")
		}

		if got := r.URL.Query().Get("per_page"); got != "100" {
			t.Fatalf("per_page query = %q, want %q", got, "100")
		}

		page := r.URL.Query().Get("page")
		switch page {
		case "", "0":
			w.Header().Set("Link", `<`+server.URL+`/orgs/test-org/repos?page=2&per_page=100&type=all>; rel="next", <`+server.URL+`/orgs/test-org/repos?page=2&per_page=100&type=all>; rel="last"`)
			writeReposResponse(t, w, "repo-one")
		case "2":
			writeReposResponse(t, w, "repo-two")
		default:
			t.Fatalf("unexpected page query %q", page)
		}
	})

	client := gogithub.NewClient(server.Client())
	baseURL := server.URL + "/"
	var err error
	client.BaseURL, err = client.BaseURL.Parse(baseURL)
	if err != nil {
		t.Fatalf("Parse(baseURL) error = %v", err)
	}

	repos, err := ListRepos(client, "test-org", "all")
	if err != nil {
		t.Fatalf("ListRepos() error = %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("ListRepos() len = %d, want %d", len(repos), 2)
	}

	if repos[0].GetName() != "repo-one" {
		t.Fatalf("repos[0].GetName() = %q, want %q", repos[0].GetName(), "repo-one")
	}

	if repos[1].GetName() != "repo-two" {
		t.Fatalf("repos[1].GetName() = %q, want %q", repos[1].GetName(), "repo-two")
	}
}

func writeReposResponse(t *testing.T, w http.ResponseWriter, names ...string) {
	t.Helper()

	repos := make([]*gogithub.Repository, 0, len(names))
	for _, name := range names {
		repos = append(repos, &gogithub.Repository{Name: gogithub.Ptr(name)})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(repos); err != nil {
		t.Fatalf("Encode(repos) error = %v", err)
	}
}
