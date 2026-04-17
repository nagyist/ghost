package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildConnectCmd(app *common.App) *cobra.Command {
	var readOnly bool

	cmd := &cobra.Command{
		Use:   "connect <name-or-id>",
		Short: "Get connection string for a database",
		Long: `Get a PostgreSQL connection string for a database.

Includes the password from ~/.pgpass if available.`,
		Example: `  # Get connection string for a database
  ghost connect my-database
  ghost connect a2x6xoj0oz

  # Get a read-only connection string
  ghost connect --read-only my-database`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch database details
			resp, err := client.GetDatabaseWithResponse(cmd.Context(), projectID, databaseRef)
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			// Handle API response
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			database := *resp.JSON200

			// Get password for database
			password, err := common.GetPassword(database, "tsdbadmin")
			if err != nil {
				cmd.PrintErrf("Warning: failed to get password: %v\n", err)
			}

			// Build and output connection string
			connStr, err := common.BuildConnectionString(common.ConnectionStringArgs{
				Database: database,
				Role:     "tsdbadmin",
				Password: password,
				ReadOnly: readOnly,
			})
			if err != nil {
				return fmt.Errorf("failed to build connection string: %w", err)
			}

			cmd.Println(connStr)
			return nil
		},
	}

	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Connect in read-only mode")

	return cmd
}
