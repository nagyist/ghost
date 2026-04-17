package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildPasswordCmd(app *common.App) *cobra.Command {
	var generate bool

	cmd := &cobra.Command{
		Use:   "password <name-or-id> [new-password]",
		Short: "Reset the password for a database",
		Long: `Reset the password for the default database user.

This changes the password on the server itself. This operation is irreversible.
Existing connections using the old password will fail to reconnect.

The new password can be provided as a positional argument, entered interactively,
or automatically generated using the --generate flag.

The password will be saved to your ~/.pgpass file for use with psql and other
PostgreSQL tools.`,
		Example: `  # Update password (interactive prompt)
  ghost password my-database

  # Update password with explicit value
  ghost password my-database "my-new-secure-password"

  # Generate a secure password
  ghost password my-database --generate`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseRef := args[0]

			// Get password from args if provided
			var password string
			if len(args) > 1 {
				password = args[1]
			}

			// Validate mutual exclusivity of password argument and --generate
			if password != "" && generate {
				return errors.New("cannot use --generate when password is provided as an argument")
			}

			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// Fetch database details
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

			// Check if the database is ready
			if err := common.CheckReady(database); err != nil {
				return handleDatabaseError(err, database.Id)
			}

			// Determine the password to use
			if generate {
				// Generate a secure password
				password, err = generateSecurePassword(32)
				if err != nil {
					return fmt.Errorf("failed to generate password: %w", err)
				}
			} else if password == "" {
				// Interactive prompt - check if we're in a terminal
				if !util.IsTerminal(cmd.InOrStdin()) {
					return errors.New("no password provided and stdin is not a terminal; use --generate or provide password as argument")
				}

				password, err = promptForPassword(cmd)
				if err != nil {
					return err
				}
			}

			// Update the password via API
			updateReq := api.UpdatePasswordRequest{Password: password}
			resp, err := client.UpdatePasswordWithResponse(cmd.Context(), projectID, database.Id, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update password: %w", err)
			}

			if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			// Save password to .pgpass
			if err := common.SavePassword(database, password, "tsdbadmin"); err != nil {
				cmd.PrintErrf("Warning: failed to save password to .pgpass: %v\n", err)
			}

			cmd.Printf("Password updated for '%s'\n", database.Name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&generate, "generate", false, "Automatically generate a secure password")

	return cmd
}

// promptForPassword prompts the user to enter a password interactively
func promptForPassword(cmd *cobra.Command) (string, error) {
	cmd.PrintErr("Enter new password: ")

	password, err := util.ReadPassword(cmd.Context(), cmd.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	cmd.PrintErrln() // newline after password entry

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	return password, nil
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	// Encode as URL-safe base64 to avoid special characters
	encoded := base64.URLEncoding.EncodeToString(bytes)

	// Trim to desired length
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return encoded, nil
}
