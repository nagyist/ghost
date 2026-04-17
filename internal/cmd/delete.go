package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildDeleteCmd(app *common.App) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:     "delete <name-or-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a database",
		Long: `Delete a database permanently.

This operation is irreversible. By default, you will be prompted to confirm
the deletion, unless you use the --confirm flag.`,
		Example: `  # Delete a database (with confirmation prompt)
  ghost delete my-database

  # Delete a database without confirmation prompt
  ghost delete my-database --confirm`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch database details to get the name
			getResp, err := client.GetDatabaseWithResponse(cmd.Context(), projectID, databaseRef)
			if err != nil {
				return fmt.Errorf("failed to get database details: %w", err)
			}

			if getResp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(getResp.StatusCode(), getResp.JSONDefault)
			}

			if getResp.JSON200 == nil {
				return errors.New("empty response from API")
			}
			database := *getResp.JSON200

			// Prompt for confirmation unless --confirm is used
			if !confirm {
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("cannot prompt for confirmation: stdin is not a terminal; use --confirm to skip")
				}
				cmd.PrintErrf("Delete '%s' (%s)? This cannot be undone. [y/N] ", database.Name, database.Id)

				confirmation, err := util.ReadLine(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read confirmation: %w", err)
				}

				confirmation = strings.ToLower(confirmation)
				if confirmation != "y" && confirmation != "yes" {
					cmd.Println("Delete operation cancelled.")
					return nil
				}
			}

			// Make the delete request using the resolved ID
			resp, err := client.DeleteDatabaseWithResponse(
				cmd.Context(),
				api.SpaceId(projectID),
				api.DatabaseRef(database.Id),
			)
			if err != nil {
				return fmt.Errorf("failed to delete database: %w", err)
			}

			// Handle response
			if resp.StatusCode() != http.StatusAccepted {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			// Remove the pgpass entry for the deleted database
			if err := common.RemovePgpassEntry(database, "tsdbadmin"); err != nil {
				cmd.PrintErrf("Warning: failed to remove .pgpass entry: %v\n", err)
			}

			cmd.Printf("Deleted '%s'\n", database.Name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}
