/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage GitHub users across organizations.",
	Long: `The user command groups user-oriented GitHub operations such as
adding a user to one or more organizations and reconciling the user's
organization, team, and cost center state.`,
}

func init() {
	rootCmd.AddCommand(userCmd)
}
