package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildSpaceCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Manage spaces",
		Long: `Manage Ghost spaces.

A space is a collection of databases with shared usage limits and billing.
The CLI operates on one space at a time — the current space. Use
'ghost space list' to see your spaces and 'ghost space use' to switch
between them.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
	}

	cmd.AddCommand(buildSpaceListCmd(app))
	cmd.AddCommand(buildSpaceUseCmd(app))
	cmd.AddCommand(buildSpaceRenameCmd(app))

	return cmd
}
