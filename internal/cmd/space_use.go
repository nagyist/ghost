package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildSpaceUseCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "use <id>",
		Aliases: []string{"switch"},
		Short:   "Switch the current space",
		Long: `Switch the current space.

Subsequent commands operate on the new current space. Run 'ghost space list'
to see your spaces and their IDs.`,
		Example: `  # Switch the current space
  ghost space use x9y8z7w6v5`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: spaceCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID := args[0]

			cfg, client, _, err := app.GetAll()
			if err != nil {
				return err
			}

			// Switching rewrites the space ID in the stored login
			// credentials. API keys are bound to a single space, so
			// switching is only possible with an OAuth login.
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

			resp, err := client.GetSpaceWithResponse(cmd.Context(), spaceID)
			if err != nil {
				return fmt.Errorf("failed to get space: %w", err)
			}
			if resp.StatusCode() == http.StatusNotFound {
				return fmt.Errorf("space '%s' not found; run 'ghost space list' to see your spaces", spaceID)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}
			space := resp.JSON200

			creds.SpaceID = space.Id
			if err := cfg.StoreCredentials(creds); err != nil {
				return fmt.Errorf("failed to store credentials: %w", err)
			}

			cmd.Printf("Switched to space '%s' (%s)\n", space.Name, space.Id)
			return nil
		},
	}

	return cmd
}
