package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
)

func buildPauseCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <name-or-id>",
		Short: "Pause a running database",
		Long:  `Pause a running database. This terminates active connections.`,
		Example: `  # Pause a database
  ghost pause my-database`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Make the pause request
			resp, err := client.PauseDatabaseWithResponse(
				cmd.Context(),
				api.SpaceId(projectID),
				api.DatabaseRef(databaseRef),
			)
			if err != nil {
				return fmt.Errorf("failed to pause database: %w", err)
			}

			// Handle API response
			if resp.StatusCode() != http.StatusAccepted {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON202 == nil {
				return errors.New("empty response from API")
			}
			database := *resp.JSON202

			cmd.Printf("Pausing '%s' (%s)...\n", database.Name, database.Id)

			return nil
		},
	}

	return cmd
}
