package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// InvoiceOutput represents the output format for a single invoice.
type InvoiceOutput struct {
	ID            string  `json:"id"`
	InvoiceNumber string  `json:"invoice_number"`
	InvoiceDate   string  `json:"invoice_date"`
	Total         float64 `json:"total"`
	Status        string  `json:"status"`
}

func buildInvoiceListCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List invoices",
		Long: `List invoices for your Ghost space.

Returns the most recent invoices.`,
		Aliases:           []string{"ls"},
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListInvoicesWithResponse(cmd.Context(), spaceID)
			if err != nil {
				return fmt.Errorf("failed to list invoices: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			invoices := resp.JSON200.Invoices

			output := make([]InvoiceOutput, len(invoices))
			for i, inv := range invoices {
				output[i] = InvoiceOutput{
					ID:            inv.ID,
					InvoiceNumber: inv.InvoiceNumber,
					InvoiceDate:   inv.InvoiceDate.Format("2006-01-02"),
					Total:         inv.Total,
					Status:        string(inv.Status),
				}
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), output)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), output)
			default:
				if len(output) == 0 {
					cmd.Println("No invoices found.")
					return nil
				}
				return outputInvoices(cmd.OutOrStdout(), output)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputInvoices(w io.Writer, invoices []InvoiceOutput) error {
	table := common.NewTable(w)

	table.Header("ID", "DATE", "TOTAL", "STATUS")
	for _, inv := range invoices {
		total := fmt.Sprintf("$%.2f", inv.Total)
		table.Append(inv.ID, inv.InvoiceDate, total, inv.Status)
	}

	return table.Render()
}
