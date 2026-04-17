package cmd

import (
	"errors"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildSQLCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql <name-or-id> [query]",
		Short: "Execute SQL query on a database",
		Long: `Execute a SQL query against a database and display the results.

If no query is provided as an argument, reads from stdin.

Multi-statement queries (semicolon-separated) are supported. Results from
all statements that return rows will be displayed.`,
		Example: `  # Select data from a table
  ghost sql my-database "SELECT * FROM users LIMIT 5"

  # Execute DDL
  ghost sql my-database "CREATE TABLE todos (id SERIAL PRIMARY KEY, title TEXT)"

  # Multi-statement query
  ghost sql my-database "INSERT INTO users (name) VALUES ('alice'); SELECT * FROM users"

  # Read query from stdin
  echo "SELECT 1" | ghost sql my-database

  # Read query from a file
  ghost sql my-database < schema.sql`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			// Get query from args or stdin
			var query string
			if len(args) > 1 {
				query = args[1]
			} else {
				if util.IsTerminal(cmd.InOrStdin()) {
					cmd.PrintErrln("Enter your query (press Ctrl+D when done):")
				}
				input, err := util.ReadAll(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read query: %w", err)
				}
				query = input
			}

			if query == "" {
				return errors.New("query cannot be empty")
			}

			cfg, client, projectID, err := app.GetAll()
			if err != nil {
				return err
			}

			// Execute the query
			result, err := common.ExecuteQuery(cmd.Context(), common.ExecuteQueryArgs{
				Client:      client,
				ProjectID:   projectID,
				DatabaseRef: databaseRef,
				Query:       query,
				Role:        "tsdbadmin",
				ReadOnly:    cfg.ReadOnly,
			})
			if err != nil {
				return handleDatabaseError(err, databaseRef)
			}

			// Display results
			return displayQueryResults(cmd, result)
		},
	}

	return cmd
}

// displayQueryResults formats and displays query results using tablewriter.
func displayQueryResults(cmd *cobra.Command, result *common.QueryResult) error {
	output := cmd.OutOrStdout()

	for i, rs := range result.ResultSets {
		// For non-SELECT queries (no columns), show the command tag
		if len(rs.Columns) == 0 {
			cmd.Println(rs.CommandTag)
			continue
		}

		// Create table with pipe-delimited style
		table := tablewriter.NewTable(output,
			tablewriter.WithHeaderAlignment(tw.AlignLeft),
			tablewriter.WithHeaderAutoFormat(tw.Off),
			tablewriter.WithRendition(tw.Rendition{
				Borders: tw.Border{
					Left:   tw.Off,
					Right:  tw.Off,
					Top:    tw.Off,
					Bottom: tw.Off,
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.Off,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.On,
					},
					Lines: tw.Lines{
						ShowHeaderLine: tw.On,
					},
				},
			}),
		)

		// Set headers from column names
		headers := make([]any, len(rs.Columns))
		for j, col := range rs.Columns {
			headers[j] = col.Name
		}
		table.Header(headers...)

		// Add rows
		for _, row := range rs.Rows {
			rowAny := make([]any, len(row))
			for j, val := range row {
				rowAny[j] = val
			}
			table.Append(rowAny...)
		}

		if err := table.Render(); err != nil {
			return fmt.Errorf("failed to render table for result set %d: %w", i, err)
		}

		if len(rs.Rows) == 1 {
			cmd.Printf("(%d row)\n\n", len(rs.Rows))
		} else {
			cmd.Printf("(%d rows)\n\n", len(rs.Rows))
		}
	}

	return nil
}
