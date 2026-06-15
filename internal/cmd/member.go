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
)

func buildMemberCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage space members",
		Long: `Manage the members of the current space.

Use 'ghost member list' to see the members of the current space,
'ghost member role' to change a member's role, and 'ghost member remove'
to remove a member from the space.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
	}

	cmd.AddCommand(buildMemberListCmd(app))
	cmd.AddCommand(buildMemberRemoveCmd(app))
	cmd.AddCommand(buildMemberRoleCmd(app))

	return cmd
}

// findMemberByEmail resolves an email address to a member of the space by
// listing the space's members. Emails are matched case-insensitively.
func findMemberByEmail(ctx context.Context, client api.ClientWithResponsesInterface, spaceID, email string) (*api.Member, error) {
	resp, err := client.ListMembersWithResponse(ctx, api.SpaceId(spaceID))
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}
	if resp.JSON200 == nil {
		return nil, errors.New("empty response from API")
	}

	for _, member := range *resp.JSON200 {
		if strings.EqualFold(member.Email, email) {
			return &member, nil
		}
	}
	return nil, fmt.Errorf("no member with email '%s' found; run 'ghost member list' to see the members of this space", email)
}
