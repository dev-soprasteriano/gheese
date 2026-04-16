/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"fmt"
	"strings"

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
		multiTransfer, err := cmd.Flags().GetBool("All")
		if err != nil {
			fmt.Println(err)
			return
		}

		source, err := cmd.Flags().GetString("Source")
		if err != nil {
			fmt.Println(err)
			return
		}

		destination, err := cmd.Flags().GetString("Destination")
		if err != nil {
			fmt.Println(err)
			return
		}

		var sourceOrg string
		var sourceRepo string
		var destOrg string
		var destRepo string

		if multiTransfer {
			sourceOrg, err = parseOrganizationReference(source, "--Source")
			if err != nil {
				fmt.Println(err)
				return
			}

			destOrg, err = parseOrganizationReference(destination, "--Destination")
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			sourceOrg, sourceRepo, err = parseRepositoryReference(source, "--Source")
			if err != nil {
				fmt.Println(err)
				return
			}

			destOrg, destRepo, err = parseRepositoryReference(destination, "--Destination")
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		c, err := github.NewClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		teamId, err := cmd.Flags().GetInt64Slice("TeamId")
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
	moveCmd.Flags().StringP("Source", "s", "", "Source reference. Use <org/repo>; with --All, use <org>.")
	moveCmd.Flags().StringP("Destination", "d", "", "Destination reference. Use <org/repo>; with --All, use <org>.")
	moveCmd.Flags().Int64SliceP("TeamId", "t", nil, "ID of team or teams to add to the repository")
	moveCmd.Flags().BoolP("All", "A", false, "This will move all the repositories for the organization.")

	if err := moveCmd.MarkFlagRequired("Source"); err != nil {
		panic(err)
	}

	if err := moveCmd.MarkFlagRequired("Destination"); err != nil {
		panic(err)
	}
}

func printBatchTransferResult(result *github.BatchTransferResult) {
	fmt.Printf("Processed %d repositories\n", result.TotalRepositories)
	fmt.Printf("Requested transfer for %d repositories\n", result.RequestedTransfers)
}

func parseRepositoryReference(input, flagName string) (string, string, error) {
	parts, err := splitReference(input, flagName)
	if err != nil {
		return "", "", err
	}

	if len(parts) != 2 {
		return "", "", fmt.Errorf("%s must use the format <org/repo>", flagName)
	}

	return parts[0], parts[1], nil
}

func parseOrganizationReference(input, flagName string) (string, error) {
	parts, err := splitReference(input, flagName)
	if err != nil {
		return "", err
	}

	if len(parts) != 1 {
		return "", fmt.Errorf("%s must use the format <org> when --All is set", flagName)
	}

	return parts[0], nil
}

func splitReference(input, flagName string) ([]string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("%s is required", flagName)
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) > 2 {
		return nil, fmt.Errorf("%s has too many segments: %q", flagName, input)
	}

	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return nil, fmt.Errorf("%s contains an empty segment: %q", flagName, input)
		}
	}

	return parts, nil
}
