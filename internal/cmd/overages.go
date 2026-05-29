package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildOveragesCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overages",
		Short: "Manage compute overages",
		Long: `Manage compute overage billing for your Ghost space.

By default, each space gets an included free compute allowance each calendar
month; when it is used up, all standard databases in the space are auto-paused
until the next month. Enabling overages lets you pay for compute beyond the
free allowance, optionally capped at a monthly compute-hour limit you choose.

Run 'ghost pricing' to see the included free allowance and the per-hour
overage rate.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
	}

	cmd.AddCommand(buildOveragesEnableCmd(app))
	cmd.AddCommand(buildOveragesDisableCmd(app))

	return cmd
}
