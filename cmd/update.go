/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"strings"

	internalgithub "github.com/dev-soprasteriano/gheese/internal/github"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "Reconcile a user back to the requested state.",
	Long:         `The update command reconciles a user to the requested organization, team, and enterprise cost center state so it can be run repeatedly in automation.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags, err := readUserCommandFlags(cmd)
		if err != nil {
			return err
		}

		req, err := buildUserSyncRequest(flags)
		if err != nil {
			return err
		}

		output := strings.ToLower(strings.TrimSpace(flags.output))

		c, err := internalgithub.NewClient()
		if err != nil {
			return err
		}

		result, syncErr := internalgithub.UpdateUser(c, req, internalgithub.UserSyncOptions{DryRun: flags.dryRun})
		if result != nil {
			if printErr := printUserSyncResult(result, output); printErr != nil {
				return printErr
			}
		}

		return syncErr
	},
}

func init() {
	userCmd.AddCommand(updateCmd)
	addUserUpdateFlags(updateCmd)
}
