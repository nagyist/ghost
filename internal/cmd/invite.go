package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildInviteCmd(app *common.App) *cobra.Command {
	var role string
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:   "invite <email>",
		Short: "Invite a user to the current space",
		Long: `Invite a user to the current space by email address.

The invitee accepts the invitation with their own Ghost account.

Roles:
  admin      Manage databases, members, and billing
  developer  Manage databases only (default)
  viewer     Read-only access`,
		Example: `  # Invite a user as a developer (the default)
  ghost invite bob@example.com

  # Invite a user as an admin
  ghost invite bob@example.com --role admin

  # List invites you've sent
  ghost invite sent`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			email := strings.ToLower(args[0])

			switch api.MemberRole(role) {
			case api.MemberRoleAdmin, api.MemberRoleDeveloper, api.MemberRoleViewer:
			case api.MemberRoleOwner:
				return errors.New("the owner role cannot be granted; every space has exactly one owner")
			default:
				return fmt.Errorf("invalid role '%s'; must be one of admin, developer, or viewer", role)
			}

			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.CreateInviteWithResponse(
				cmd.Context(),
				api.SpaceId(spaceID),
				api.CreateInviteJSONRequestBody{
					Email: email,
					Role:  new(api.MemberRole(role)),
				},
			)
			if err != nil {
				return fmt.Errorf("failed to create invite: %w", err)
			}
			if resp.StatusCode() != http.StatusCreated {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON201 == nil {
				return errors.New("empty response from API")
			}

			created := resp.JSON201

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), created)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), created)
			default:
				spaceName := currentSpaceName(cmd.Context(), client, spaceID)
				outputInviteCreated(cmd, *created, spaceID, spaceName)
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&role, "role", string(api.MemberRoleDeveloper), "Role to grant the invitee (admin|developer|viewer)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	if err := cmd.RegisterFlagCompletionFunc("role", cobra.FixedCompletions(grantableRoles, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	cmd.AddCommand(buildInviteListCmd(app))
	cmd.AddCommand(buildInviteSentCmd(app))
	cmd.AddCommand(buildInviteCancelCmd(app))
	cmd.AddCommand(buildInviteReceivedCmd(app))
	cmd.AddCommand(buildInviteAcceptCmd(app))
	cmd.AddCommand(buildInviteDeclineCmd(app))

	return cmd
}

func outputInviteCreated(cmd *cobra.Command, invite api.Invite, spaceID, spaceName string) {
	// Show the space as "name (id)", matching `ghost usage`; fall back to just
	// the ID when the name couldn't be resolved.
	space := spaceID
	if spaceName != "" {
		space = fmt.Sprintf("%s (%s)", spaceName, spaceID)
	}
	// "admin" is the only role that takes "an".
	article := "a"
	if invite.Role == api.MemberRoleAdmin {
		article = "an"
	}
	cmd.Printf("Invited %s to space %s as %s %s.\n\n", invite.Email, space, article, invite.Role)
	cmd.Println("If they're new to Ghost, the invitation will be accepted automatically")
	cmd.Println("at signup. If they already use Ghost, they can accept the invite with")
	cmd.Printf("'ghost invite accept %s'.\n\n", spaceID)
	cmd.Printf("Note: the invitation is tied to %s; they must log in to\n", invite.Email)
	cmd.Println("Ghost with a GitHub account with that primary email.")
}

// currentSpaceName resolves the display name of the current space, best-effort.
// It returns an empty string if the name can't be resolved, so callers can fall
// back to showing just the space ID.
func currentSpaceName(ctx context.Context, client api.ClientWithResponsesInterface, spaceID string) string {
	resp, err := client.ListSpacesWithResponse(ctx)
	if err != nil || resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return ""
	}
	for _, s := range *resp.JSON200 {
		if s.Id == spaceID {
			return s.Name
		}
	}
	return ""
}
