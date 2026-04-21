/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"encoding/json"
	"fmt"

	internalgithub "github.com/dev-soprasteriano/gheese/internal/github"
)

func printUserSyncResult(result *internalgithub.UserSyncResult, output string) error {
	switch output {
	case "json":
		encoded, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}

		fmt.Println(string(encoded))
		return nil
	case "text":
		fmt.Printf("Mode: %s\n", result.Mode)
		if result.Login != "" {
			fmt.Printf("User: %s\n", result.Login)
		}
		if result.Email != "" {
			fmt.Printf("Email: %s\n", result.Email)
		}

		if result.CostCenter != nil {
			fmt.Printf("Cost center: %s (%s)\n", result.CostCenter.Name, result.CostCenter.Action)
			if result.CostCenter.PreviousCostCenter != "" {
				fmt.Printf("Previous cost center: %s\n", result.CostCenter.PreviousCostCenter)
			}
		}

		for _, org := range result.Organizations {
			fmt.Printf("Organization %s: %s\n", org.Organization, org.Membership)
			for _, team := range org.Teams {
				fmt.Printf("  Team %s: %s", team.Team, team.Action)
				if team.State != "" {
					fmt.Printf(" (%s)", team.State)
				}
				if team.Error != "" {
					fmt.Printf(" - %s", team.Error)
				}
				fmt.Println()
			}
			if org.Error != "" {
				fmt.Printf("  Error: %s\n", org.Error)
			}
		}

		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, err := range result.Errors {
				fmt.Printf("- %s\n", err)
			}
		}

		return nil
	default:
		return fmt.Errorf("invalid output %q: use text or json", output)
	}
}
