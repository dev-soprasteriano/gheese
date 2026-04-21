/*
Copyright © 2026 Sopra Steria AS

This file is part of gheese and is licensed under the GNU General Public License v3.0.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	internalgithub "github.com/dev-soprasteriano/gheese/internal/github"
	"github.com/spf13/cobra"
)

const defaultUserInvitationRole = "direct_member"

type userCommandFlags struct {
	enterprise string
	login      string
	email      string
	costCenter string
	role       string
	file       string
	output     string
	orgs       []string
	teams      []string
	dryRun     bool
}

func addUserAddFlags(cmd *cobra.Command) {
	cmd.Flags().String("enterprise", "", "Enterprise slug that owns the cost center")
	cmd.Flags().String("login", "", "GitHub login for the user to invite")
	cmd.Flags().String("email", "", "Email address for the user. Add can use email without login")
	cmd.Flags().String("role", defaultUserInvitationRole, "Organization role for invitations: direct_member, admin, or billing_manager")
	cmd.Flags().String("file", "", "Path to a JSON file containing the requested user state")
	cmd.Flags().String("output", "text", "Output format: text or json")
	cmd.Flags().StringArray("org", nil, "Organization to include in the requested state")
	cmd.Flags().StringArray("team", nil, "Team to include in the requested state, using the format <org>/<team-slug>")
	cmd.Flags().Bool("dry-run", false, "Validate and report actions without changing GitHub")
}

func addUserUpdateFlags(cmd *cobra.Command) {
	cmd.Flags().String("enterprise", "", "Enterprise slug that owns the cost center")
	cmd.Flags().String("login", "", "GitHub login for the user to reconcile")
	cmd.Flags().String("email", "", "Optional email address for reporting and invite matching")
	cmd.Flags().String("cost-center", "", "Required enterprise cost center name")
	cmd.Flags().String("role", defaultUserInvitationRole, "Organization role to keep in request files")
	cmd.Flags().String("file", "", "Path to a JSON file containing the requested user state")
	cmd.Flags().String("output", "text", "Output format: text or json")
	cmd.Flags().StringArray("org", nil, "Organization to include in the requested state")
	cmd.Flags().StringArray("team", nil, "Team to include in the requested state, using the format <org>/<team-slug>")
	cmd.Flags().Bool("dry-run", false, "Validate and report actions without changing GitHub")
}

func readUserCommandFlags(cmd *cobra.Command) (userCommandFlags, error) {
	var flags userCommandFlags
	var err error

	flags.enterprise, err = cmd.Flags().GetString("enterprise")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.login, err = cmd.Flags().GetString("login")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.email, err = cmd.Flags().GetString("email")
	if err != nil {
		return userCommandFlags{}, err
	}

	if cmd.Flags().Lookup("cost-center") != nil {
		flags.costCenter, err = cmd.Flags().GetString("cost-center")
		if err != nil {
			return userCommandFlags{}, err
		}
	}

	if cmd.Flags().Lookup("role") != nil {
		flags.role, err = cmd.Flags().GetString("role")
		if err != nil {
			return userCommandFlags{}, err
		}
	}

	flags.file, err = cmd.Flags().GetString("file")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.output, err = cmd.Flags().GetString("output")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.orgs, err = cmd.Flags().GetStringArray("org")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.teams, err = cmd.Flags().GetStringArray("team")
	if err != nil {
		return userCommandFlags{}, err
	}

	flags.dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return userCommandFlags{}, err
	}

	return flags, nil
}

func buildUserSyncRequest(flags userCommandFlags) (internalgithub.UserSyncRequest, error) {
	if strings.TrimSpace(flags.output) != "" && !slices.Contains([]string{"json", "text"}, strings.ToLower(strings.TrimSpace(flags.output))) {
		return internalgithub.UserSyncRequest{}, fmt.Errorf("invalid output %q: use text or json", flags.output)
	}

	if strings.TrimSpace(flags.file) != "" {
		if hasDirectUserInput(flags) {
			return internalgithub.UserSyncRequest{}, fmt.Errorf("request flags cannot be combined with --file")
		}

		return loadUserSyncRequestFromFile(flags.file)
	}

	orgMemberships := make([]internalgithub.UserOrganizationAssignment, 0, len(flags.orgs))
	orgIndexes := make(map[string]int, len(flags.orgs))
	for _, org := range flags.orgs {
		orgName := strings.TrimSpace(org)
		if orgName == "" {
			return internalgithub.UserSyncRequest{}, fmt.Errorf("--org cannot contain an empty value")
		}

		orgKey := strings.ToLower(orgName)
		if _, exists := orgIndexes[orgKey]; exists {
			return internalgithub.UserSyncRequest{}, fmt.Errorf("organization %q is duplicated", orgName)
		}

		orgIndexes[orgKey] = len(orgMemberships)
		orgMemberships = append(orgMemberships, internalgithub.UserOrganizationAssignment{Name: orgName})
	}

	for _, teamReference := range flags.teams {
		orgName, teamSlug, err := parseTeamReference(teamReference)
		if err != nil {
			return internalgithub.UserSyncRequest{}, err
		}

		orgIndex, exists := orgIndexes[strings.ToLower(orgName)]
		if !exists {
			return internalgithub.UserSyncRequest{}, fmt.Errorf("team %q references organization %q, which is not present in --org", teamReference, orgName)
		}

		orgMemberships[orgIndex].Teams = append(orgMemberships[orgIndex].Teams, teamSlug)
	}

	return internalgithub.UserSyncRequest{
		Enterprise:    flags.enterprise,
		Login:         flags.login,
		Email:         flags.email,
		CostCenter:    flags.costCenter,
		Role:          flags.role,
		Organizations: orgMemberships,
	}, nil
}

func loadUserSyncRequestFromFile(path string) (internalgithub.UserSyncRequest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return internalgithub.UserSyncRequest{}, fmt.Errorf("read request file: %w", err)
	}

	var req internalgithub.UserSyncRequest
	if err := json.Unmarshal(content, &req); err != nil {
		return internalgithub.UserSyncRequest{}, fmt.Errorf("parse request file: %w", err)
	}

	return req, nil
}

func hasDirectUserInput(flags userCommandFlags) bool {
	return strings.TrimSpace(flags.enterprise) != "" ||
		strings.TrimSpace(flags.login) != "" ||
		strings.TrimSpace(flags.email) != "" ||
		strings.TrimSpace(flags.costCenter) != "" ||
		(strings.TrimSpace(flags.role) != "" && strings.TrimSpace(flags.role) != defaultUserInvitationRole) ||
		len(flags.orgs) > 0 ||
		len(flags.teams) > 0
}

func parseTeamReference(input string) (string, string, error) {
	parts, err := splitReference(input, "--team")
	if err != nil {
		return "", "", err
	}

	if len(parts) != 2 {
		return "", "", fmt.Errorf("--team must use the format <org/team-slug>")
	}

	return parts[0], parts[1], nil
}
