package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
	"github.com/timescale/ghost/internal/util"
)

func buildSpaceLeaveCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave the current space",
		Long: `Leave the current space, removing yourself from its members.

You cannot leave a space you own. By default, you will be prompted to
confirm, unless you use the --confirm flag.`,
		Example: `  # Leave the current space (with confirmation prompt)
  ghost space leave

  # Leave without confirmation prompt
  ghost space leave --confirm`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, spaceID, err := app.GetAll()
			if err != nil {
				return err
			}

			// Resolve the space name for the confirmation prompt and messages.
			spaceResp, err := client.GetSpaceWithResponse(cmd.Context(), spaceID)
			if err != nil {
				return fmt.Errorf("failed to get space: %w", err)
			}
			if spaceResp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(spaceResp.StatusCode(), spaceResp.JSONDefault)
			}
			if spaceResp.JSON200 == nil {
				return errors.New("empty response from API")
			}
			space := spaceResp.JSON200

			// Prompt for confirmation unless --confirm is used.
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Leave space '%s' (%s)? [y/N] ", space.Name, space.ID)

				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				confirmation = strings.ToLower(confirmation)
				if confirmation != "y" && confirmation != "yes" {
					cmd.Println("Leave operation cancelled.")
					return nil
				}
			}

			resp, err := client.LeaveSpaceWithResponse(cmd.Context(), api.SpaceID(spaceID))
			if err != nil {
				return fmt.Errorf("failed to leave space: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			result := resp.JSON200
			cmd.Printf("Left space '%s' (%s)\n", result.SpaceName, result.SpaceID)

			// The current space now points at one we no longer belong to, which
			// would break subsequent commands. Switch back to the user's owned
			// space — every user has exactly one, and login defaults to it.
			switchToOwnedSpace(cmd, cfg, client)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}

// switchToOwnedSpace points the stored current space back at the user's owned
// space after they leave a joined one. It is best-effort: on any failure it
// prints a hint to switch manually rather than failing the leave, which has
// already succeeded.
func switchToOwnedSpace(cmd *cobra.Command, cfg *config.Config, client api.ClientWithResponsesInterface) {
	hint := func() {
		cmd.Println("Run 'ghost space use <id>' to select a space for subsequent commands.")
	}

	listResp, err := client.ListSpacesWithResponse(cmd.Context())
	if err != nil || listResp.JSON200 == nil {
		hint()
		return
	}

	var owned *api.Space
	for i, space := range *listResp.JSON200 {
		if space.Role != nil && *space.Role == api.MemberRoleOwner {
			owned = &(*listResp.JSON200)[i]
			break
		}
	}
	if owned == nil {
		hint()
		return
	}

	creds, err := cfg.GetCredentials()
	if err != nil {
		hint()
		return
	}
	creds.SpaceID = owned.ID
	if err := cfg.StoreCredentials(creds); err != nil {
		hint()
		return
	}

	cmd.Printf("Switched to space '%s' (%s)\n", owned.Name, owned.ID)
}
