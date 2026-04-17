package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// ApiKeyCreateOutput represents the output format for the api-key create command.
type ApiKeyCreateOutput struct {
	ApiKey string `json:"api_key"`
}

func buildApiKeyCreateCmd(app *common.App) *cobra.Command {
	var name string
	var jsonOutput bool
	var yamlOutput bool
	var envOutput bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API key",
		Long: `Create a new API key for your Ghost space.

The API key is only shown once — make sure to save it.
API keys can be used to authenticate with Ghost by setting the
GHOST_API_KEY environment variable.`,
		Example: `  # Create an API key with auto-generated name
  ghost api-key create

  # Create an API key with a custom name
  ghost api-key create --name "CI/CD Key"

  # Output as environment variables (useful for .env files)
  ghost api-key create --env > .env

  # Output as JSON
  ghost api-key create --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, projectID, err := app.GetClient()
			if err != nil {
				return err
			}

			// If no name provided, generate a default based on the user's name
			if name == "" {
				name, err = defaultApiKeyName(cmd.Context(), client)
				if err != nil {
					return err
				}
			}

			// Create the API key
			resp, err := client.CreateApiKeyWithResponse(cmd.Context(), projectID, api.CreateApiKeyJSONRequestBody{
				Name: name,
			})
			if err != nil {
				return fmt.Errorf("failed to create API key: %w", err)
			}

			if resp.StatusCode() != http.StatusCreated {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			if resp.JSON201 == nil {
				return errors.New("empty response from API")
			}

			output := ApiKeyCreateOutput{
				ApiKey: resp.JSON201.ApiKey,
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), output)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), output)
			case envOutput:
				cmd.Printf("GHOST_API_KEY=%s\n", output.ApiKey)
			default:
				cmd.Printf("Created API key '%s'\n", name)
				cmd.Printf("API key: %s\n", output.ApiKey)
				cmd.Println()
				cmd.Println("This key will not be shown again. Make sure to save it.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "API key name (auto-generated if not provided)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.Flags().BoolVar(&envOutput, "env", false, "Output as environment variables")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml", "env")

	return cmd
}

// defaultApiKeyName fetches the current user's name from the API and returns
// a default API key name like "<user>'s API Key".
func defaultApiKeyName(ctx context.Context, client api.ClientWithResponsesInterface) (string, error) {
	resp, err := client.AuthInfoWithResponse(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user info: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user info: %s", resp.Status())
	}

	if resp.JSON200 == nil {
		return "", errors.New("empty response from API")
	}

	authInfo := resp.JSON200

	var userName string
	switch authInfo.Type {
	case api.AuthInfoTypeUser:
		if authInfo.User != nil {
			userName = authInfo.User.Name
		}
	case api.AuthInfoTypeApiKey:
		if authInfo.ApiKey != nil {
			userName = authInfo.ApiKey.UserName
		}
	}

	if userName != "" {
		return fmt.Sprintf("%s's API Key", userName), nil
	}

	return "My API Key", nil
}
