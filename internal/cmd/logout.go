package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
)

func buildLogoutCmd(app *common.App) *cobra.Command {
	return &cobra.Command{
		Use:               "logout",
		Short:             "Remove stored credentials",
		Long:              `Remove stored API key and clear authentication credentials.`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, _, err := app.GetAll()
			if err != nil {
				return err
			}

			creds, err := cfg.GetCredentials()
			if err != nil {
				return fmt.Errorf("failed to read credentials: %w", err)
			}

			// If the user has OAuth token credentials, revoke the refresh token
			// server-side before removing local credentials. Legacy API key users
			// skip this step since the endpoint only accepts JWT auth (API keys
			// don't have a refresh token to revoke)
			if creds.Token != nil && creds.Token.RefreshToken != "" {
				resp, err := client.LogoutWithResponse(cmd.Context(), api.LogoutJSONRequestBody{
					RefreshToken: creds.Token.RefreshToken,
				})
				if err != nil {
					return fmt.Errorf("failed to revoke refresh token: %w", err)
				} else if statusCode := resp.StatusCode(); statusCode != http.StatusNoContent {
					return fmt.Errorf("failed to revoke refresh token: unexpected status code: %d", statusCode)
				}
			}

			if err := cfg.RemoveCredentials(); err != nil {
				return fmt.Errorf("failed to remove credentials: %w", err)
			}

			cmd.Println("Successfully logged out")
			return nil
		},
	}
}
