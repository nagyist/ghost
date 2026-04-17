package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// PaymentMethodOutput represents the output format for a payment method
type PaymentMethodOutput struct {
	ID              string `json:"id"`
	Brand           string `json:"brand"`
	Last4           string `json:"last4"`
	ExpMonth        int    `json:"exp_month"`
	ExpYear         int    `json:"exp_year"`
	Primary         bool   `json:"is_primary"`
	PendingDeletion bool   `json:"pending_deletion"`
}

func outputPaymentMethods(w io.Writer, methods []PaymentMethodOutput) error {
	table := tablewriter.NewTable(w,
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
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

	table.Header("ID", "BRAND", "LAST4", "EXPIRES", "PRIMARY", "PENDING DELETION")
	for _, pm := range methods {
		expires := fmt.Sprintf("%02d/%d", pm.ExpMonth, pm.ExpYear)
		primary := "no"
		if pm.Primary {
			primary = "yes"
		}
		pendingDeletion := "no"
		if pm.PendingDeletion {
			pendingDeletion = "yes"
		}
		table.Append(pm.ID, pm.Brand, pm.Last4, expires, primary, pendingDeletion)
	}

	return table.Render()
}

func buildPaymentListCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:               "list",
		Short:             "List payment methods",
		Long:              `List payment methods for your Ghost space.`,
		Args:              cobra.NoArgs,
		Aliases:           []string{"ls"},
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListPaymentMethodsWithResponse(cmd.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to list payment methods: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			methods := resp.JSON200.PaymentMethods
			if len(methods) == 0 {
				cmd.Println("No payment methods on file.")
				cmd.Println("Run 'ghost payment add' to add a payment method.")
				return nil
			}

			output := make([]PaymentMethodOutput, len(methods))
			for i, pm := range methods {
				output[i] = PaymentMethodOutput{
					ID:              pm.Id,
					Brand:           pm.Brand,
					Last4:           pm.Last4,
					ExpMonth:        pm.ExpMonth,
					ExpYear:         pm.ExpYear,
					Primary:         pm.Primary,
					PendingDeletion: pm.PendingDeletion,
				}
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), output)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), output)
			default:
				return outputPaymentMethods(cmd.OutOrStdout(), output)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}
