package github

import (
	"fmt"
	"os"

	"github.com/google/go-github/v84/github"
)

func NewClient() (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("The environment variable GITHUB_TOKEN is not set.")
	}
	client := github.NewClient(nil).WithAuthToken(token)

	return client, nil
}
