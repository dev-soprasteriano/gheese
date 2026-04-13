package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v84/github"
)

func ListRepos(c *github.Client, org, visibility string) ([]*github.Repository, error) {
	filter, err := normalizeVisibilityFilter(visibility)
	if err != nil {
		return nil, err
	}

	options := &github.RepositoryListByOrgOptions{
		Type: filter,
	}

	repos, _, err := c.Repositories.ListByOrg(context.Background(), org, options)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func normalizeVisibilityFilter(visibility string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(visibility)) {
	case "", "all":
		return "all", nil
	case "public":
		return "public", nil
	case "private":
		return "private", nil
	default:
		return "", fmt.Errorf("invalid visibility %q: use all, public, or private", visibility)
	}
}
