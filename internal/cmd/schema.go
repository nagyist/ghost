package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildSchemaCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema <name-or-id>",
		Short: "Display database schema information",
		Long: `Display database schema information including tables, views, materialized views,
and enum types with their columns, constraints, and indexes.`,
		Example:           `  ghost schema my-database`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			schema, err := common.FetchDatabaseSchema(cmd.Context(), common.FetchDatabaseSchemaArgs{
				Client:      client,
				ProjectID:   projectID,
				DatabaseRef: databaseRef,
			})
			if err != nil {
				return handleDatabaseError(err, databaseRef)
			}

			cmd.Print(common.FormatSchema(schema))
			return nil
		},
	}

	return cmd
}
