package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildPsqlCmd(app *common.App) *cobra.Command {
	var readOnly bool

	cmd := &cobra.Command{
		Use:   "psql <name-or-id> [-- <psql-flags>...]",
		Short: "Connect to a database using psql",
		Long: `Connect to a database using psql.

The psql client must already be installed on your machine. The database
password is read from ~/.pgpass. If no password is found, the command
will fail with an error.

Any flags after -- are passed directly to psql.`,
		Example: `  # Connect to a database
  ghost psql my-database

  # Pass additional psql flags
  ghost psql my-database -- --single-transaction --quiet`,
		Args:              cobra.ArbitraryArgs,
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Separate database ref from additional psql flags
			dbArgs, psqlFlags := splitArgsAtDash(cmd, args)
			if len(dbArgs) != 1 {
				return errors.New("exactly one database name or ID is required")
			}
			databaseRef := dbArgs[0]

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch database details
			resp, err := client.GetDatabaseWithResponse(cmd.Context(), projectID, databaseRef)
			if err != nil {
				return fmt.Errorf("failed to get database: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			database := *resp.JSON200

			// Check if the database is ready to accept connections
			if err := common.CheckReady(database); err != nil {
				return handleDatabaseError(err, database.Id)
			}

			// Get password for database
			password, err := common.GetPassword(database, "tsdbadmin")
			if err != nil {
				if errors.Is(err, common.ErrPasswordNotFound) {
					return fmt.Errorf("password not found\n\nRun 'ghost password %s' to reset the password, or add the entry to ~/.pgpass manually", database.Id)
				}
				return fmt.Errorf("failed to retrieve password: %w", err)
			}

			// Build connection string without password (password is passed via env var)
			connStr, err := common.BuildConnectionString(common.ConnectionStringArgs{
				Database: database,
				Role:     "tsdbadmin",
				ReadOnly: readOnly,
			})
			if err != nil {
				return fmt.Errorf("failed to build connection string: %w", err)
			}

			// Find psql
			psqlPath, err := exec.LookPath("psql")
			if err != nil {
				return errors.New("psql not found in PATH; please install PostgreSQL client tools")
			}

			// Launch psql with PGPASSWORD set
			psqlArgs := append([]string{connStr}, psqlFlags...)
			psqlCmd := exec.CommandContext(cmd.Context(), psqlPath, psqlArgs...)
			// Unwrap the underlying *os.File from the command's writers.
			// os/exec only passes file descriptors directly to the child
			// process when the writer is an *os.File; any other io.Writer
			// causes it to create a pipe, which breaks TTY detection in
			// psql and suppresses interactive output (prompts, banners, etc.).
			psqlCmd.Stdin = cmd.InOrStdin()
			psqlCmd.Stdout = util.TryUnwrapFile(cmd.OutOrStdout())
			psqlCmd.Stderr = util.TryUnwrapFile(cmd.ErrOrStderr())
			psqlCmd.Env = append(os.Environ(), "PGPASSWORD="+password)

			return psqlCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Connect in read-only mode")

	return cmd
}

// splitArgsAtDash separates positional args from flags after "--" using Cobra's ArgsLenAtDash.
func splitArgsAtDash(cmd *cobra.Command, args []string) (positional []string, extra []string) {
	n := cmd.ArgsLenAtDash()
	if n >= 0 {
		return args[:n], args[n:]
	}
	return args, nil
}
