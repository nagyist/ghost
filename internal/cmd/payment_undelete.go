package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildPaymentUndeleteCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelete <payment-id>",
		Short: "Cancel a pending payment method deletion",
		Long:  `Cancel the pending deletion of a payment method.`,
		Example: `  # Cancel deletion for a specific payment method
  ghost payment undelete pm_xxx`,
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
			if !pm.PendingDeletion {
				return errors.New("payment method does not have a pending deletion")
			}

			// Cancel the pending deletion
			resp, err := client.CancelPaymentMethodDeletionWithResponse(cmd.Context(), projectID, args[0])
			if err != nil {
				return fmt.Errorf("failed to cancel payment method deletion: %w", err)
			}

			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Printf("Cancelled pending deletion for %s ending in %s.\n", pm.Brand, pm.Last4)
			return nil
		},
	}

	return cmd
}
