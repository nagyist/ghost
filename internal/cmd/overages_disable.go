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

func buildOveragesDisableCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable compute overages",
		Long: `Disable compute overage billing for your Ghost space.

After disabling, the compute limit resets to the included free allowance.
If your current month-to-date usage is already above that, your non-dedicated
databases will be paused.`,
		Example: `  # Disable overages (prompts for confirmation)
  ghost overages disable

  # Disable without confirmation prompt
  ghost overages disable --confirm`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Disable compute overages? Standard databases will pause once usage reaches the included free allowance, or pause immediately if you are already above that. [y/N] ")
				answer, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if a := strings.ToLower(answer); a != "y" && a != "yes" {
					cmd.Println("Disable cancelled.")
					return nil
				}
			}

			resp, err := client.UpdateOveragesWithResponse(cmd.Context(), projectID, api.UpdateOverageSettingsRequest{
				Enabled: false,
			})
			if err != nil {
				return fmt.Errorf("failed to disable overages: %w", err)
			}
			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Println("Overages disabled.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
