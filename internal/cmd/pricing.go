package cmd

import (
	"fmt"
	"io"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildPricingCmd(app *common.App) *cobra.Command {
	var jsonOutput, yamlOutput bool

	cmd := &cobra.Command{
		Use:               "pricing",
		Aliases:           []string{"price", "prices"},
		Short:             "Show dedicated database pricing",
		Long:              `Show pricing for dedicated databases, including compute pricing for each available size and the storage rate.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The /pricing endpoint is public, but we reuse the standard
			// authenticated client builder for simplicity — login is cheap
			// and there's no existing unauthenticated-client path.
			client, _, err := app.GetClient()
			if err != nil {
				return err
			}

			output, err := common.FetchPricing(cmd.Context(), client)
			if err != nil {
				return err
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), output)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), output)
			default:
				return renderPricingText(cmd.OutOrStdout(), output)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")
	return cmd
}

func renderPricingText(w io.Writer, p common.PricingOutput) error {
	fmt.Fprintln(w, lipgloss.NewStyle().Bold(true).Render("Dedicated"))
	table := tablewriter.NewTable(w,
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		// Disable auto-formatting so "$/HOUR" isn't split into "$ / HOUR" on
		// non-alphanumeric boundaries.
		tablewriter.WithHeaderAutoFormat(tw.Off),
		tablewriter.WithPadding(tw.Padding{Left: "", Right: "  ", Overwrite: true}),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Left:   tw.Off,
				Right:  tw.Off,
				Top:    tw.Off,
				Bottom: tw.Off,
			},
			Settings: tw.Settings{
				Separators: tw.Separators{
					ShowHeader:     tw.Off,
					ShowFooter:     tw.Off,
					BetweenRows:    tw.Off,
					BetweenColumns: tw.Off,
				},
				Lines: tw.Lines{
					ShowHeaderLine: tw.Off,
				},
			},
		}),
	)
	table.Header("SIZE", "VCPU", "MEMORY", "$/HOUR", "$/MONTH")
	for _, c := range p.Dedicated.Compute {
		table.Append(
			c.Size,
			fmt.Sprintf("%.1f", float64(c.MilliCPU)/1000),
			fmt.Sprintf("%d GiB", c.MemoryGiB),
			fmt.Sprintf("$%.4f", c.PricePerHour),
			fmt.Sprintf("$%.2f", c.PricePerMonth),
		)
	}
	if err := table.Render(); err != nil {
		return err
	}

	storage := p.Dedicated.Storage
	fmt.Fprintf(w, "\nStorage: first %d GiB per database included; $%.6f/GiB/hour ($%.2f/GiB/month) above that.\n",
		storage.IncludedGiBPerDatabase, storage.PricePerGiBHour, storage.PricePerGiBMonth)
	return nil
}
