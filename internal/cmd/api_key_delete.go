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

func buildApiKeyDeleteCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <prefix>",
		Short: "Delete an API key",
		Long:  `Delete an API key from your Ghost space.`,
		Example: `  # Delete an API key
  ghost api-key delete gt_abc123

  # Delete without confirmation prompt
  ghost api-key delete gt_abc123 --confirm`,
		Aliases:           []string{"rm"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: apiKeyPrefixCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			prefix := args[0]

			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Delete API key with prefix '%s'? [y/N] ", prefix)
				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				if c := strings.ToLower(confirmation); c != "y" && c != "yes" {
					cmd.Println("Delete cancelled.")
					return nil
				}
			}

			resp, err := client.DeleteApiKeyWithResponse(cmd.Context(), projectID, prefix)
			if err != nil {
				return fmt.Errorf("failed to delete API key: %w", err)
			}

			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Printf("Deleted API key with prefix '%s'.\n", prefix)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
