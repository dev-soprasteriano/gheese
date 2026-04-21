package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	gogithub "github.com/google/go-github/v84/github"
)

const defaultOrgRole = "direct_member"

// UserOrganizationAssignment declares the desired state for one organization.
type UserOrganizationAssignment struct {
	Name  string   `json:"name"`
	Teams []string `json:"teams,omitempty"`
}

// UserSyncRequest declares the desired state for a user.
type UserSyncRequest struct {
	Enterprise    string                       `json:"enterprise"`
	Login         string                       `json:"login"`
	Email         string                       `json:"email,omitempty"`
	CostCenter    string                       `json:"cost_center"`
	Role          string                       `json:"role,omitempty"`
	Organizations []UserOrganizationAssignment `json:"organizations"`
}

// UserSyncOptions controls how a user sync is executed.
type UserSyncOptions struct {
	DryRun bool
	Mode   string
}

// CostCenterAssignmentResult describes the outcome of cost center reconciliation.
type CostCenterAssignmentResult struct {
	Name               string `json:"name"`
	Action             string `json:"action"`
	PreviousCostCenter string `json:"previous_cost_center,omitempty"`
}

// TeamSyncResult describes the outcome for one team.
type TeamSyncResult struct {
	Team   string `json:"team"`
	Action string `json:"action"`
	State  string `json:"state,omitempty"`
	Error  string `json:"error,omitempty"`
}

// OrganizationSyncResult describes the outcome for one organization.
type OrganizationSyncResult struct {
	Organization string           `json:"organization"`
	Membership   string           `json:"membership"`
	Teams        []TeamSyncResult `json:"teams,omitempty"`
	Error        string           `json:"error,omitempty"`
}

// UserSyncResult reports the outcome of a user add or update run.
type UserSyncResult struct {
	Mode          string                      `json:"mode"`
	DryRun        bool                        `json:"dry_run"`
	Login         string                      `json:"login,omitempty"`
	Email         string                      `json:"email,omitempty"`
	CostCenter    *CostCenterAssignmentResult `json:"cost_center,omitempty"`
	Organizations []OrganizationSyncResult    `json:"organizations"`
	Errors        []string                    `json:"errors,omitempty"`
}

type resolvedOrganizationAssignment struct {
	Name    string
	Teams   []string
	TeamIDs []int64
}

type preparedUserSync struct {
	Request           UserSyncRequest
	Organizations     []resolvedOrganizationAssignment
	CurrentCostCenter string
	CostCenterID      string
	InviteeID         int64
}

// NormalizeUserSyncRequest validates and normalizes a requested user state.
func NormalizeUserSyncRequest(req UserSyncRequest) (UserSyncRequest, error) {
	req.Enterprise = strings.TrimSpace(req.Enterprise)
	req.Login = strings.TrimSpace(req.Login)
	req.Email = strings.TrimSpace(req.Email)
	req.CostCenter = strings.TrimSpace(req.CostCenter)
	req.Role = strings.TrimSpace(req.Role)

	if req.Role == "" {
		req.Role = defaultOrgRole
	}

	if req.Enterprise == "" {
		return UserSyncRequest{}, fmt.Errorf("enterprise is required")
	}

	if req.Login == "" && req.Email == "" {
		return UserSyncRequest{}, fmt.Errorf("login or email is required")
	}

	if !slices.Contains([]string{"admin", "billing_manager", "direct_member"}, req.Role) {
		return UserSyncRequest{}, fmt.Errorf("invalid role %q: use admin, billing_manager, or direct_member", req.Role)
	}

	if len(req.Organizations) == 0 {
		return UserSyncRequest{}, fmt.Errorf("at least one organization is required")
	}

	seenOrgs := make(map[string]struct{}, len(req.Organizations))
	normalizedOrgs := make([]UserOrganizationAssignment, 0, len(req.Organizations))
	for _, org := range req.Organizations {
		orgName := strings.TrimSpace(org.Name)
		if orgName == "" {
			return UserSyncRequest{}, fmt.Errorf("organization name is required")
		}

		orgKey := strings.ToLower(orgName)
		if _, exists := seenOrgs[orgKey]; exists {
			return UserSyncRequest{}, fmt.Errorf("organization %q is duplicated", orgName)
		}
		seenOrgs[orgKey] = struct{}{}

		seenTeams := make(map[string]struct{}, len(org.Teams))
		normalizedTeams := make([]string, 0, len(org.Teams))
		for _, team := range org.Teams {
			teamSlug := strings.TrimSpace(team)
			if teamSlug == "" {
				return UserSyncRequest{}, fmt.Errorf("organization %q contains an empty team", orgName)
			}

			teamKey := strings.ToLower(teamSlug)
			if _, exists := seenTeams[teamKey]; exists {
				return UserSyncRequest{}, fmt.Errorf("team %q is duplicated in organization %q", teamSlug, orgName)
			}
			seenTeams[teamKey] = struct{}{}
			normalizedTeams = append(normalizedTeams, teamSlug)
		}

		normalizedOrgs = append(normalizedOrgs, UserOrganizationAssignment{
			Name:  orgName,
			Teams: normalizedTeams,
		})
	}

	req.Organizations = normalizedOrgs

	return req, nil
}

// AddUser creates the initial access state for a user by sending invitations.
func AddUser(c *gogithub.Client, req UserSyncRequest, opts UserSyncOptions) (*UserSyncResult, error) {
	opts.Mode = "add"
	if strings.TrimSpace(req.CostCenter) != "" {
		return nil, fmt.Errorf("cost center is only supported by update")
	}
	return reconcileUser(c, req, opts)
}

// UpdateUser brings an existing user back to the requested state.
func UpdateUser(c *gogithub.Client, req UserSyncRequest, opts UserSyncOptions) (*UserSyncResult, error) {
	opts.Mode = "update"
	if strings.TrimSpace(req.Login) == "" {
		return nil, fmt.Errorf("login is required for update")
	}
	if strings.TrimSpace(req.CostCenter) == "" {
		return nil, fmt.Errorf("cost center is required for update")
	}
	return reconcileUser(c, req, opts)
}

func reconcileUser(c *gogithub.Client, req UserSyncRequest, opts UserSyncOptions) (*UserSyncResult, error) {
	if c == nil {
		return nil, fmt.Errorf("github client is nil")
	}

	normalized, err := NormalizeUserSyncRequest(req)
	if err != nil {
		return nil, err
	}

	prepared, err := prepareUserSync(context.Background(), c, normalized)
	if err != nil {
		return nil, err
	}

	result := &UserSyncResult{
		Mode:          opts.Mode,
		DryRun:        opts.DryRun,
		Login:         prepared.Request.Login,
		Email:         prepared.Request.Email,
		Organizations: make([]OrganizationSyncResult, 0, len(prepared.Organizations)),
	}
	if prepared.Request.CostCenter != "" {
		result.CostCenter = &CostCenterAssignmentResult{Name: prepared.Request.CostCenter}
	}

	var syncErrors []error
	for _, org := range prepared.Organizations {
		orgResult := OrganizationSyncResult{
			Organization: org.Name,
		}

		membership, pendingInvite, membershipErr := getOrganizationState(context.Background(), c, prepared.Request, org.Name)
		if membershipErr != nil {
			orgResult.Error = membershipErr.Error()
			result.Errors = append(result.Errors, orgResult.Error)
			result.Organizations = append(result.Organizations, orgResult)
			syncErrors = append(syncErrors, fmt.Errorf("inspect organization %q: %w", org.Name, membershipErr))
			continue
		}

		createdInvite := false
		switch {
		case membership != nil && membership.GetState() == "active":
			orgResult.Membership = "already-active"
		case membership != nil && membership.GetState() == "pending":
			orgResult.Membership = "membership-pending"
		case pendingInvite:
			orgResult.Membership = "invite-pending"
		case opts.DryRun:
			orgResult.Membership = "would-invite"
			createdInvite = true
		default:
			inviteOpts := &gogithub.CreateOrgInvitationOptions{
				Role:   gogithub.Ptr(prepared.Request.Role),
				TeamID: slices.Clone(org.TeamIDs),
			}
			if prepared.InviteeID != 0 {
				inviteOpts.InviteeID = gogithub.Ptr(prepared.InviteeID)
			} else {
				inviteOpts.Email = gogithub.Ptr(prepared.Request.Email)
			}

			_, _, inviteErr := c.Organizations.CreateOrgInvitation(context.Background(), org.Name, inviteOpts)
			if inviteErr != nil {
				orgResult.Error = inviteErr.Error()
				result.Errors = append(result.Errors, orgResult.Error)
				result.Organizations = append(result.Organizations, orgResult)
				syncErrors = append(syncErrors, fmt.Errorf("create invitation for organization %q: %w", org.Name, inviteErr))
				continue
			}

			orgResult.Membership = "invite-created"
			createdInvite = true
		}

		orgResult.Teams = make([]TeamSyncResult, 0, len(org.Teams))
		for index, team := range org.Teams {
			teamResult := TeamSyncResult{Team: team}

			switch {
			case opts.DryRun && createdInvite:
				teamResult.Action = "queued-with-invite"
			case createdInvite:
				teamResult.Action = "queued-with-invite"
			case prepared.Request.Login == "":
				teamResult.Action = "pending-login"
			case opts.DryRun:
				teamResult.Action = "would-ensure"
			default:
				teamMembership, _, teamErr := c.Teams.AddTeamMembershipBySlug(
					context.Background(),
					org.Name,
					team,
					prepared.Request.Login,
					&gogithub.TeamAddTeamMembershipOptions{Role: "member"},
				)
				if teamErr != nil {
					teamResult.Action = "error"
					teamResult.Error = teamErr.Error()
					result.Errors = append(result.Errors, fmt.Sprintf("ensure team %q in organization %q: %s", team, org.Name, teamErr.Error()))
					syncErrors = append(syncErrors, fmt.Errorf("ensure team %q in organization %q: %w", team, org.Name, teamErr))
				} else {
					teamResult.State = teamMembership.GetState()
					if teamMembership.GetState() == "pending" {
						teamResult.Action = "invited"
					} else {
						teamResult.Action = "ensured"
					}
				}
			}

			if teamResult.Action == "" {
				teamResult.Action = "queued-with-invite"
			}

			orgResult.Teams = append(orgResult.Teams, teamResult)

			if teamResult.Error != "" && orgResult.Error == "" {
				orgResult.Error = fmt.Sprintf("team reconciliation failed for %d team(s)", len(org.Teams[index:]))
			}
		}

		result.Organizations = append(result.Organizations, orgResult)
	}

	if result.CostCenter != nil {
		switch {
		case prepared.CurrentCostCenter == prepared.Request.CostCenter:
			result.CostCenter.Action = "already-assigned"
		case opts.DryRun && prepared.CurrentCostCenter != "":
			result.CostCenter.Action = "would-reassign"
			result.CostCenter.PreviousCostCenter = prepared.CurrentCostCenter
		case opts.DryRun:
			result.CostCenter.Action = "would-assign"
		case prepared.CurrentCostCenter != "":
			added, _, addErr := c.Enterprise.AddResourcesToCostCenter(
				context.Background(),
				prepared.Request.Enterprise,
				prepared.CostCenterID,
				gogithub.CostCenterResourceRequest{Users: []string{prepared.Request.Login}},
			)
			if addErr != nil {
				result.CostCenter.Action = "error"
				result.Errors = append(result.Errors, fmt.Sprintf("assign cost center: %s", addErr.Error()))
				syncErrors = append(syncErrors, fmt.Errorf("assign cost center: %w", addErr))
			} else {
				result.CostCenter.Action = "reassigned"
				for _, reassigned := range added.ReassignedResources {
					if strings.EqualFold(reassigned.GetName(), prepared.Request.Login) {
						result.CostCenter.PreviousCostCenter = reassigned.GetPreviousCostCenter()
						break
					}
				}
				if result.CostCenter.PreviousCostCenter == "" {
					result.CostCenter.PreviousCostCenter = prepared.CurrentCostCenter
				}
			}
		default:
			_, _, addErr := c.Enterprise.AddResourcesToCostCenter(
				context.Background(),
				prepared.Request.Enterprise,
				prepared.CostCenterID,
				gogithub.CostCenterResourceRequest{Users: []string{prepared.Request.Login}},
			)
			if addErr != nil {
				result.CostCenter.Action = "error"
				result.Errors = append(result.Errors, fmt.Sprintf("assign cost center: %s", addErr.Error()))
				syncErrors = append(syncErrors, fmt.Errorf("assign cost center: %w", addErr))
			} else {
				result.CostCenter.Action = "assigned"
			}
		}
	}

	if len(syncErrors) > 0 {
		return result, errors.Join(syncErrors...)
	}

	return result, nil
}

func prepareUserSync(ctx context.Context, c *gogithub.Client, req UserSyncRequest) (*preparedUserSync, error) {
	var inviteeID int64
	if req.Login != "" {
		user, _, err := c.Users.Get(ctx, req.Login)
		if err != nil {
			return nil, fmt.Errorf("get user %q: %w", req.Login, err)
		}
		inviteeID = user.GetID()
	}

	var costCenterID string
	var currentCostCenter string
	if req.CostCenter != "" {
		costCenters, _, err := c.Enterprise.ListCostCenters(ctx, req.Enterprise, nil)
		if err != nil {
			return nil, fmt.Errorf("list cost centers: %w", err)
		}

		for _, center := range costCenters.CostCenters {
			if strings.EqualFold(center.Name, req.CostCenter) {
				costCenterID = center.ID
			}

			for _, resource := range center.Resources {
				if req.Login != "" && strings.EqualFold(resource.Type, "User") && strings.EqualFold(resource.Name, req.Login) {
					currentCostCenter = center.Name
				}
			}
		}

		if costCenterID == "" {
			return nil, fmt.Errorf("cost center %q was not found in enterprise %q", req.CostCenter, req.Enterprise)
		}
	}

	resolvedOrgs := make([]resolvedOrganizationAssignment, 0, len(req.Organizations))
	for _, org := range req.Organizations {
		resolvedOrg := resolvedOrganizationAssignment{
			Name:  org.Name,
			Teams: slices.Clone(org.Teams),
		}

		for _, teamSlug := range org.Teams {
			team, _, teamErr := c.Teams.GetTeamBySlug(ctx, org.Name, teamSlug)
			if teamErr != nil {
				return nil, fmt.Errorf("resolve team %q in organization %q: %w", teamSlug, org.Name, teamErr)
			}

			resolvedOrg.TeamIDs = append(resolvedOrg.TeamIDs, team.GetID())
		}

		resolvedOrgs = append(resolvedOrgs, resolvedOrg)
	}

	return &preparedUserSync{
		Request:           req,
		Organizations:     resolvedOrgs,
		CurrentCostCenter: currentCostCenter,
		CostCenterID:      costCenterID,
		InviteeID:         inviteeID,
	}, nil
}

func getOrganizationState(ctx context.Context, c *gogithub.Client, req UserSyncRequest, org string) (*gogithub.Membership, bool, error) {
	if req.Login != "" {
		membership, resp, err := c.Organizations.GetOrgMembership(ctx, req.Login, org)
		if err == nil {
			return membership, false, nil
		}

		if !isNotFound(resp, err) {
			return nil, false, err
		}
	}

	pendingInvites, inviteResp, inviteErr := c.Organizations.ListPendingOrgInvitations(ctx, org, &gogithub.ListOptions{PerPage: 100})
	if inviteErr != nil {
		return nil, false, inviteErr
	}

	for inviteResp != nil && inviteResp.NextPage != 0 {
		options := &gogithub.ListOptions{PerPage: 100, Page: inviteResp.NextPage}
		nextInvites, nextResp, nextErr := c.Organizations.ListPendingOrgInvitations(ctx, org, options)
		if nextErr != nil {
			return nil, false, nextErr
		}
		pendingInvites = append(pendingInvites, nextInvites...)
		inviteResp = nextResp
	}

	for _, invitation := range pendingInvites {
		if strings.EqualFold(invitation.GetLogin(), req.Login) {
			return nil, true, nil
		}

		if req.Email != "" && strings.EqualFold(invitation.GetEmail(), req.Email) {
			return nil, true, nil
		}
	}

	return nil, false, nil
}

func isNotFound(resp *gogithub.Response, err error) bool {
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return true
	}

	var rateLimitErr *gogithub.ErrorResponse
	if errors.As(err, &rateLimitErr) && rateLimitErr.Response != nil && rateLimitErr.Response.StatusCode == http.StatusNotFound {
		return true
	}

	return false
}
