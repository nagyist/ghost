package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildPaymentAddCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "add",
		Short:             "Add a payment method",
		Long:              `Add a payment method (credit card) by opening a secure payment page in your browser.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, projectID, err := app.GetAll()
			if err != nil {
				return err
			}

			resp, err := client.CreatePaymentMethodSetupWithResponse(cmd.Context(), projectID)
			if err != nil {
				return fmt.Errorf("failed to create payment setup: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			paymentURL := cfg.APIURL + resp.JSON200.PaymentUrl

			cmd.Println("Opening browser to add payment method...")
			if err := common.OpenBrowser(paymentURL); err != nil {
				cmd.PrintErrln("Could not open browser automatically.")
				cmd.PrintErrf("Please open this URL in your browser:\n\n  %s\n\n", paymentURL)
			}

			cmd.Println("Complete the payment form in your browser.")
			return nil
		},
	}

	return cmd
}
