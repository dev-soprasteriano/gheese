package github

import (
	"context"
	"github.com/google/go-github/v84/github"
)

func ListRepos(c *github.Client, org string) ([]*github.Repository, error) {
	repos, _, err := c.Repositories.ListByOrg(context.Background(), org, nil)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
