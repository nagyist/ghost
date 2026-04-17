package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildPaymentPrimaryCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "primary <payment-id>",
		Short: "Set the primary payment method",
		Long:  `Set a payment method as the primary payment method for your space.`,
		Example: `  # Set a specific payment method as primary
  ghost payment primary pm_xxx`,
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
			if pm.Primary {
				return errors.New("payment method is already primary")
			}

			// Set as primary
			resp, err := client.SetPaymentMethodPrimaryWithResponse(cmd.Context(), projectID, args[0])
			if err != nil {
				return fmt.Errorf("failed to set primary payment method: %w", err)
			}

			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Printf("%s ending in %s is now your primary payment method.\n", pm.Brand, pm.Last4)
			return nil
		},
	}

	return cmd
}
