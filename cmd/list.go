/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"fmt"

	"github.com/dev-soprasteriano/gheese/internal/github"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Get all repositories owned by an organization",
	Long: `This command lists all of the available repositories from an organization.
	
This is run in the context of the users token, set with the environment variable GITHUB_TOKEN. In other words, it will only list
the repositories the user has access to see.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := github.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		repos, err := github.ListRepos(c, args[0])
		if err != nil {
			fmt.Println(err)
		}

		for _, repo := range repos {
			fmt.Println(repo.GetName())
		}
	},
}

func init() {
	repoCmd.AddCommand(listCmd)
}
