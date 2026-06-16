package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
	"github.com/timescale/ghost/internal/util"
)

func buildInviteAcceptCmd(app *common.App) *cobra.Command {
	var switchSpace bool

	cmd := &cobra.Command{
		Use:   "accept <space-id>",
		Short: "Accept an invitation",
		Long: `Accept an invitation you've received, joining the space.

The space is identified by its ID, as shown by 'ghost invite received'. After
joining, you can switch the CLI's current space to the new space. By default
you'll be prompted; use --switch or --switch=false to decide without a prompt.`,
		Example: `  # Accept an invitation
  ghost invite accept x9y8z7w6v5

  # Accept and immediately switch to the new space
  ghost invite accept x9y8z7w6v5 --switch`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: inviteReceivedCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID := args[0]

			cfg, client, _, err := app.GetAll()
			if err != nil {
				return err
			}

			resp, err := client.AcceptInviteWithResponse(cmd.Context(), api.SpaceId(spaceID))
			if err != nil {
				return fmt.Errorf("failed to accept invitation: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			joined := resp.JSON200
			cmd.Printf("Joined space '%s' (%s)\n", joined.SpaceName, joined.SpaceId)

			// Decide whether to switch the current space to the joined one.
			doSwitch := switchSpace
			if !cmd.Flags().Changed("switch") {
				if !util.IsTerminal(cmd.InOrStdin()) {
					cmd.Printf("Run 'ghost space use %s' to switch to it.\n", joined.SpaceId)
					return nil
				}
				cmd.PrintErrf("Switch to this space now? [Y/n] ")
				answer, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				answer = strings.ToLower(strings.TrimSpace(answer))
				doSwitch = answer == "" || answer == "y" || answer == "yes"
			}

			if !doSwitch {
				cmd.Printf("Run 'ghost space use %s' to switch to it.\n", joined.SpaceId)
				return nil
			}

			if err := setCurrentSpace(cfg, joined.SpaceId); err != nil {
				return err
			}
			cmd.Printf("Switched to space '%s' (%s)\n", joined.SpaceName, joined.SpaceId)
			return nil
		},
	}

	cmd.Flags().BoolVar(&switchSpace, "switch", false, "Switch the current space to the joined space")

	return cmd
}

// setCurrentSpace rewrites the stored credentials' space ID to the given space.
// Switching is only possible with an OAuth login, since API keys are bound to a
// single space.
func setCurrentSpace(cfg *config.Config, spaceID string) error {
	if os.Getenv("GHOST_API_KEY") != "" {
		return errors.New("cannot switch spaces when authenticated with an API key; unset GHOST_API_KEY and run 'ghost login'")
	}
	creds, err := cfg.GetCredentials()
	if err != nil {
		return fmt.Errorf("failed to read credentials: %w", err)
	}
	if creds.Token == nil {
		return errors.New("cannot switch spaces when authenticated with an API key; run 'ghost login'")
	}
	creds.SpaceID = spaceID
	if err := cfg.StoreCredentials(creds); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}
	return nil
}
