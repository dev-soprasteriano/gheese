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

var addCmd = &cobra.Command{
	Use:          "add",
	Aliases:      []string{"onboard"},
	Short:        "Invite a user to one or more organizations.",
	Long:         `The add command creates organization invitations and queues the requested team assignments for a user across one or more organizations.`,
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

		result, syncErr := internalgithub.AddUser(c, req, internalgithub.UserSyncOptions{DryRun: flags.dryRun})
		if result != nil {
			if printErr := printUserSyncResult(result, output); printErr != nil {
				return printErr
			}
		}

		return syncErr
	},
}

func init() {
	userCmd.AddCommand(addCmd)
	addUserAddFlags(addCmd)
}
