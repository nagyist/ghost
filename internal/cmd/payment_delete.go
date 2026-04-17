package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildPaymentDeleteCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <payment-id>",
		Short: "Delete a payment method",
		Long:  `Delete a payment method.`,
		Example: `  # Delete a specific payment method
  ghost payment delete pm_xxx

  # Delete without confirmation prompt
  ghost payment delete pm_xxx --confirm`,
		Aliases:           []string{"rm"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: paymentMethodIDCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch payment method
			getResp, err := client.GetPaymentMethodWithResponse(cmd.Context(), projectID, args[0])
			if err != nil {
				return fmt.Errorf("failed to get payment method: %w", err)
			}

			if getResp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(getResp.StatusCode(), getResp.JSONDefault)
			}

			if getResp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			pm := getResp.JSON200

			// Confirm deletion
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Delete %s ending in %s? [y/N] ", pm.Brand, pm.Last4)
				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if c := strings.ToLower(confirmation); c != "y" && c != "yes" {
					cmd.Println("Delete cancelled.")
					return nil
				}
			}

			// Delete
			delResp, err := client.DeletePaymentMethodWithResponse(cmd.Context(), projectID, args[0])
			if err != nil {
				return fmt.Errorf("failed to delete payment method: %w", err)
			}

			if delResp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(delResp.StatusCode(), delResp.JSONDefault)
			}

			cmd.Printf("Deleted %s ending in %s.\n", pm.Brand, pm.Last4)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
