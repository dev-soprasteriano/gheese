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

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Transfer repositories to another organization.",
	Long: `The move command transfers a single repository or, with --All, walks
through all repositories in a source organization and lets you confirm each
transfer before it is requested from GitHub.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := github.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		sourceOrg, err := cmd.Flags().GetString("SourceOrganization")
		if err != nil {
			fmt.Println(err)
			return
		}

		sourceRepo, err := cmd.Flags().GetString("SourceRepository")
		if err != nil {
			fmt.Println(err)
			return
		}

		destOrg, err := cmd.Flags().GetString("DestinationOrganization")
		if err != nil {
			fmt.Println(err)
			return
		}

		destRepo, err := cmd.Flags().GetString("DestinationRepository")
		if err != nil {
			fmt.Println(err)
			return
		}

		teamId, err := cmd.Flags().GetInt64Slice("TeamId")
		if err != nil {
			fmt.Println(err)
			return
		}

		multiTransfer, err := cmd.Flags().GetBool("All")
		if err != nil {
			fmt.Println(err)
			return
		}

		if multiTransfer {
			result, err := github.TransferMultipleRepo(c, sourceOrg, destOrg, teamId)
			if err != nil {
				fmt.Println(err)
				return
			}

			printBatchTransferResult(result)
			return
		} else {
			repo, res, err := github.TransferRepo(
				c,
				sourceOrg,
				sourceRepo,
				destOrg,
				&destRepo,
				teamId,
			)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(repo)
			fmt.Println("---")
			fmt.Println(res)
		}
	},
}

func init() {
	repoCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringP("SourceOrganization", "o", "", "Where does the repository currently live?")
	moveCmd.Flags().StringP("SourceRepository", "r", "", "Which repository do you want to move?")
	moveCmd.Flags().StringP("DestinationOrganization", "O", "", "Which organization will you transfer the repo to?")
	moveCmd.Flags().StringP("DestinationRepository", "R", "", "What do you want to name the repo after the transfer?")
	moveCmd.Flags().Int64SliceP("TeamId", "t", nil, "ID of team or teams to add to the repository")
	moveCmd.Flags().BoolP("All", "A", false, "This will move all the repositories for the organization.")
}

func printBatchTransferResult(result *github.BatchTransferResult) {
	fmt.Printf("Processed %d repositories\n", result.TotalRepositories)
	fmt.Printf("Requested transfer for %d repositories\n", result.RequestedTransfers)
}
