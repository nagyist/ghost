package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildInviteCancelCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:     "cancel <email>",
		Aliases: []string{"revoke", "rm"},
		Short:   "Cancel an invite you've sent",
		Long: `Cancel the pending invite sent to an email address for the current space.

By default, you will be prompted to confirm, unless you use the --confirm flag.`,
		Example: `  # Cancel an invite (with confirmation prompt)
  ghost invite cancel bob@example.com

  # Cancel an invite without a confirmation prompt
  ghost invite cancel bob@example.com --confirm`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: inviteSentCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			email := strings.ToLower(args[0])

			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Prompt for confirmation unless --confirm is used
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Cancel the invite to %s? [y/N] ", email)

				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				confirmation = strings.ToLower(confirmation)
				if confirmation != "y" && confirmation != "yes" {
					cmd.Println("Cancel operation aborted.")
					return nil
				}
			}

			resp, err := client.CancelInviteWithResponse(
				cmd.Context(),
				api.SpaceId(spaceID),
				api.InviteEmail(email),
			)
			if err != nil {
				return fmt.Errorf("failed to cancel invite: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			cmd.Printf("Cancelled the invite to %s\n", resp.JSON200.Email)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
