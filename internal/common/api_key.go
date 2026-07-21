package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/timescale/ghost/internal/api"
)

// DefaultAPIKeyName fetches the current user's name from the API and returns
// a default API key name like "<user>'s API Key".
func DefaultAPIKeyName(ctx context.Context, client api.ClientWithResponsesInterface) (string, error) {
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
