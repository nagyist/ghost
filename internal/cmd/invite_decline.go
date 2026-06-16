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

func buildInviteDeclineCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "decline <space-id>",
		Short: "Decline an invitation",
		Long: `Decline an invitation you've received.

The space is identified by its ID, as shown by 'ghost invite received'. By
default, you will be prompted to confirm, unless you use the --confirm flag.`,
		Example: `  # Decline an invitation (with confirmation prompt)
  ghost invite decline x9y8z7w6v5

  # Decline without a confirmation prompt
  ghost invite decline x9y8z7w6v5 --confirm`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: inviteReceivedCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID := args[0]

			client, _, err := app.GetClient()
			if err != nil {
				return err
			}

			invitation, err := findReceivedInvite(cmd.Context(), client, spaceID)
			if err != nil {
				return err
			}

			// Prompt for confirmation unless --confirm is used
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Decline the invitation to '%s'? [y/N] ", invitation.SpaceName)

				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				confirmation = strings.ToLower(confirmation)
				if confirmation != "y" && confirmation != "yes" {
					cmd.Println("Decline operation aborted.")
					return nil
				}
			}

			resp, err := client.DeclineInviteWithResponse(cmd.Context(), api.SpaceId(spaceID))
			if err != nil {
				return fmt.Errorf("failed to decline invitation: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			cmd.Printf("Declined the invitation to '%s'\n", resp.JSON200.SpaceName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}

// findReceivedInvite resolves a space ID to one of the caller's pending received
// invites by listing them. A user has at most one pending invite per space, so
// the space ID identifies it unambiguously.
func findReceivedInvite(ctx context.Context, client api.ClientWithResponsesInterface, spaceID string) (*api.ReceivedInvite, error) {
	resp, err := client.ListReceivedInvitesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}
	if resp.JSON200 == nil {
		return nil, errors.New("empty response from API")
	}

	for _, invitation := range *resp.JSON200 {
		if invitation.SpaceId == spaceID {
			return &invitation, nil
		}
	}
	return nil, fmt.Errorf("no pending invitation to space '%s' found; run 'ghost invite received' to see your invitations", spaceID)
}
