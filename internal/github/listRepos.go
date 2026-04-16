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
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allRepos []*github.Repository

	for {
		repos, resp, err := c.Repositories.ListByOrg(context.Background(), org, options)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos...)

		if resp == nil || resp.NextPage == 0 {
			break
		}

		options.Page = resp.NextPage
	}

	return allRepos, nil
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
