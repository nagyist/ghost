package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/timescale/ghost/internal/analytics"
	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/config"
)

var (
	// Cache of validated API Keys. Useful for avoiding unnecessary calls to the
	// /auth/info and /analytics/identify endpoints when the API client is
	// loaded multiple times using credentials provided via the GHOST_API_KEY
	// env var (e.g. when using the MCP server, which re-fetches the API client
	// for each tool call).
	validatedAPIKeyCache = map[string]*api.AuthInfo{}
)

// newAPIClient initializes a [api.ClientWithResponses] and returns it along
// with the current project ID. Credentials are pulled from the environment (if
// present), or loaded from storage (either the keyring or fallback file). When
// pulled from the environment, the credentials are first validated by hitting
// the /auth/info endpoint (which also allows us to fetch the project ID), and
// the user is identified for the sake of analytics by hitting the /analytics/identify
// endpoint. When credentials are pulled from storage, those operations should
// have already been performed during authentication.
func newAPIClient(ctx context.Context, cfg *config.Config) (api.ClientWithResponsesInterface, string, error) {
	// Credentials in the environment take priority
	apiKeyEnv := os.Getenv("GHOST_API_KEY")

	// If there were no credentials in the environment, try to load stored credentials
	if apiKeyEnv == "" {
		creds, err := cfg.GetCredentials()
		if err != nil {
			return nil, "", ExitWithCode(ExitAuthenticationError, fmt.Errorf("authentication required: %w", err))
		}

		// Select the appropriate auth method based on credential type
		var auth api.AuthMethod
		switch {
		case creds.Token != nil:
			token, err := refreshTokenIfNeeded(ctx, cfg, creds.Token, creds.ProjectID)
			if err != nil {
				return nil, "", ExitWithCode(ExitAuthenticationError, err)
			}

			auth = api.TokenAuth(token)
		case creds.APIKey != "":
			// API key auth (legacy login flow)
			auth = api.LegacyAPIKeyAuth(creds.APIKey)
		default:
			return nil, "", ExitWithCode(ExitAuthenticationError, errors.New("authentication required: no valid credentials found"))
		}

		// Create API client
		client, err := api.NewGhostClient(cfg.APIURL, auth)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create API client: %w", err)
		}

		// Return immediately. Credentials were already verified and user was
		// already identified for analytics during authentication.
		return client, creds.ProjectID, nil
	}

	// Create API client using environment variable credentials
	client, err := api.NewGhostClient(cfg.APIURL, api.APIKeyAuth(apiKeyEnv))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create API client: %w", err)
	}

	// Check whether this API Key has already been validated, and use the
	// cached auth info if so. Otherwise, validate it.
	authInfo, ok := validatedAPIKeyCache[apiKeyEnv]
	if !ok {
		// Validate the API key and identify the user for analytics
		authInfo, err = identifyWithAPIKey(ctx, cfg, client)
		if err != nil {
			return nil, "", fmt.Errorf("API key validation failed: %w", err)
		}
		validatedAPIKeyCache[apiKeyEnv] = authInfo
	}

	return client, authInfo.ApiKey.SpaceId, nil
}

func refreshTokenIfNeeded(ctx context.Context, cfg *config.Config, token *oauth2.Token, projectID string) (*oauth2.Token, error) {
	// OAuth token auth (new login flow). Get a valid token via the
	// token source, which will refresh automatically if expired.
	tokenSource := newRefreshTokenSource(ctx, cfg, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		// If the refresh token is expired or revoked, the server
		// returns "invalid_grant". Treat this as a "not logged in"
		// error so the user is prompted to log in again.
		if re, ok := errors.AsType[*oauth2.RetrieveError](err); ok && re.ErrorCode == "invalid_grant" {
			return nil, errors.New("authentication required: session expired")
		}
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Persist the refreshed token so subsequent invocations can reuse it
	if newToken.AccessToken != token.AccessToken {
		if err := cfg.StoreCredentials(config.Credentials{
			Token:     newToken,
			ProjectID: projectID,
		}); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}
	}

	return newToken, nil
}

// newRefreshTokenSource creates an oauth2.TokenSource that automatically refreshes
// the access token using the refresh token when it expires. The token is
// considered expired 15 minutes before its actual expiry to ensure commands
// that start near the end of a token's lifetime don't fail mid-execution.
func newRefreshTokenSource(ctx context.Context, cfg *config.Config, token *oauth2.Token) oauth2.TokenSource {
	oauthCfg := oauth2.Config{
		ClientID: ghostClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL:  cfg.APIURL + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
	return oauth2.ReuseTokenSourceWithExpiry(token, oauthCfg.TokenSource(ctx, token), 15*time.Minute)
}

// identifyWithAPIKey calls the /auth/info endpoint to validate an API key and
// identify the user for analytics. The response includes the space ID associated
// with the API key.
func identifyWithAPIKey(ctx context.Context, cfg *config.Config, client api.ClientWithResponsesInterface) (*api.AuthInfo, error) {
	// Call the /auth/info endpoint to validate credentials and get auth info
	resp, err := client.AuthInfoWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	// Check the response status
	if resp.StatusCode() != 200 {
		if resp.JSONDefault != nil {
			return nil, resp.JSONDefault
		}
		return nil, fmt.Errorf("unexpected API response: %d", resp.StatusCode())
	}

	if resp.JSON200 == nil {
		return nil, errors.New("empty response from API")
	}

	authInfo := resp.JSON200

	// Identify the user with analytics
	a := analytics.New(cfg, client, authInfo.ApiKey.SpaceId)
	a.Identify(
		analytics.Property("userId", authInfo.ApiKey.UserId),
		analytics.Property("email", authInfo.ApiKey.UserEmail),
	)

	return authInfo, nil
}
