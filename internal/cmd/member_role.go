package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
)

func buildMemberRoleCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role <email> <admin|developer|viewer>",
		Short: "Change a member's role",
		Long: `Change a member's role in the current space.

Roles:
  admin      Manage databases, members, and billing
  developer  Manage databases only
  viewer     Read-only access

The owner role cannot be granted; every space has exactly one owner, and
the owner's role cannot be changed.`,
		Example: `  # Make a member an admin
  ghost member role bob@example.com admin

  # Restrict a member to read-only access
  ghost member role bob@example.com viewer`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: memberEmailRoleCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]
			role := strings.ToLower(args[1])

			switch role {
			case "admin", "developer", "viewer":
			case "owner":
				return errors.New("the owner role cannot be granted; every space has exactly one owner")
			default:
				return fmt.Errorf("invalid role '%s'; must be one of admin, developer, or viewer", role)
			}

			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			member, err := findMemberByEmail(cmd.Context(), client, spaceID, email)
			if err != nil {
				return err
			}

			resp, err := client.UpdateMemberRoleWithResponse(
				cmd.Context(),
				api.SpaceId(spaceID),
				api.MemberUserId(member.UserId),
				api.UpdateMemberRoleRequest{Role: api.MemberRole(role)},
			)
			if err != nil {
				return fmt.Errorf("failed to update member role: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			updated := resp.JSON200
			cmd.Printf("Changed role of %s (%s) to %s\n", updated.Name, updated.Email, updated.Role)
			return nil
		},
	}

	return cmd
}
