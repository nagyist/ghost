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

func buildOveragesEnableCmd(app *common.App) *cobra.Command {
	var limit int64
	var confirm bool

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable compute overages",
		Long: `Enable compute overage billing for your Ghost space.

Once enabled, you will be charged for compute beyond the included free
allowance each calendar month (see 'ghost pricing'). By default there is no
monthly cap on overage usage — pass --limit <hours> to set one. When the cap
is reached, standard databases in the space auto-pause until the next month.

A payment method must be on file before overages can be enabled. Run
'ghost payment add' to add one.

This command is also used to update an existing overages-enabled space:
re-run it with a different --limit value (or with no flag, to switch to
no-limit mode).`,
		Example: `  # Enable overages with a 200-hour monthly cap
  ghost overages enable --limit 200

  # Enable overages with no monthly cap (charges have no upper bound)
  ghost overages enable

  # Skip the no-limit confirmation prompt
  ghost overages enable --confirm`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// An invalid --limit (at or below the free-tier allowance) is
			// rejected server-side, so no client-side validation is needed here.
			limitSet := cmd.Flags().Changed("limit")

			// Confirm before enabling overages with no monthly limit, since
			// the user is opting into unbounded billing exposure.
			if !limitSet && !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; pass --limit <hours> or --confirm to skip")
				}
				cmd.PrintErrf("You are enabling overages with no monthly limit. Your databases will never be auto-paused for hitting a compute limit, and you will be billed for all overage usage with no upper bound. Continue? [y/N] ")
				answer, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}
				if a := strings.ToLower(answer); a != "y" && a != "yes" {
					cmd.Println("Enable cancelled.")
					return nil
				}
			}

			req := api.UpdateOverageSettingsRequest{Enabled: true}
			if limitSet {
				req.ComputeLimitMinutes = new(limit * 60)
			}

			resp, err := client.UpdateOveragesWithResponse(cmd.Context(), projectID, req)
			if err != nil {
				return fmt.Errorf("failed to enable overages: %w", err)
			}
			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if limitSet {
				cmd.Printf("Overages enabled. You will be charged for compute beyond the included free allowance, up to %d hours/month. See 'ghost pricing' for current rates.\n", limit)
			} else {
				cmd.Println("Overages enabled with no monthly limit. You will be charged for ALL compute usage beyond the included free allowance, with no upper bound. See 'ghost pricing' for current rates.")
			}
			return nil
		},
	}

	cmd.Flags().Int64Var(&limit, "limit", 0, "Monthly compute cap in hours (must exceed the included free allowance). Omit for no cap.")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip the no-limit confirmation prompt")

	return cmd
}
