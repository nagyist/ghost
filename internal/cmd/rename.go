package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
)

func buildRenameCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <name-or-id> <new-name>",
		Short: "Rename a database",
		Long:  `Rename a database.`,
		Example: `  # Rename a database
  ghost rename my-database my-new-name`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]
			newName := args[1]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch database details to resolve name/ID
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

			resp, err := client.RenameDatabaseWithResponse(
				cmd.Context(),
				api.SpaceId(projectID),
				api.DatabaseRef(database.Id),
				api.RenameDatabaseRequest{Name: newName},
			)
			if err != nil {
				return fmt.Errorf("failed to rename database: %w", err)
			}

			if resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Printf("Renamed '%s' (%s) to '%s'\n", database.Name, database.Id, newName)

			return nil
		},
	}

	return cmd
}
