package github

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/go-github/v84/github"
)

type UserCostCenter struct {
	Name       string
	CostCenter string
}

func GetUsersMissingCC(c *github.Client, ent string) ([]UserCostCenter, error) {
	context := context.Background()

	l, _, err := c.Enterprise.GetConsumedLicenses(context, ent, nil)
	if err != nil {
		return nil, fmt.Errorf("get consumed licenses: %w", err)
	}

	cc, _, err := c.Enterprise.ListCostCenters(context, ent, nil)
	if err != nil {
		return nil, fmt.Errorf("list cost centers: %w", err)
	}

	resourceCostCenters := make(map[string]string)
	for _, center := range cc.CostCenters {
		for _, resource := range center.Resources {
			resourceCostCenters[resource.Name] = center.Name
		}
	}

	users := make([]UserCostCenter, 0, len(l.Users))
	for _, u := range l.Users {
		costcenter := resourceCostCenters[u.GithubComLogin]
		if costcenter == "" {
			costcenter = "none"
		}

		users = append(users, UserCostCenter{
			Name:       u.GithubComLogin,
			CostCenter: costcenter,
		})
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	return users, nil
}
