package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildInvoiceCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "invoice",
		Short:             "View invoices",
		Long:              `View invoices for your Ghost space.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
	}

	cmd.AddCommand(buildInvoiceListCmd(app))
	cmd.AddCommand(buildInvoiceViewCmd(app))

	return cmd
}
