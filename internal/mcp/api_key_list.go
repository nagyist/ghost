package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// APIKeyListInput represents input for ghost_api_key_list (no parameters).
type APIKeyListInput struct{}

func (APIKeyListInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[APIKeyListInput](nil))
}

// APIKeyListOutput represents output for ghost_api_key_list. The API's own
// api.APIKey type is used directly — it exposes only the prefix, name, and
// creation time, never the secret.
type APIKeyListOutput struct {
	APIKeys []api.APIKey `json:"api_keys"`
}

func (APIKeyListOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[APIKeyListOutput](nil))
	items := schema.Properties["api_keys"].Items
	items.Properties["prefix"].Description = "Stable prefix identifier for the API key (starts with 'gt_'), identifying the key without exposing the secret"
	items.Properties["name"].Description = "User-provided label for the API key"
	items.Properties["created_at"].Description = "Time the API key was created"
	return schema
}

func newAPIKeyListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_api_key_list",
		Title:        "List API Keys",
		Description:  "List all API keys in the current Ghost space.",
		InputSchema:  APIKeyListInput{}.Schema(),
		OutputSchema: APIKeyListOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(false),
			Title:         "List API Keys",
		},
	}
}

func (s *Server) handleAPIKeyList(ctx context.Context, req *mcp.CallToolRequest, input APIKeyListInput) (*mcp.CallToolResult, APIKeyListOutput, error) {
	client, spaceID, err := s.app.GetClient()
	if err != nil {
		return nil, APIKeyListOutput{}, err
	}

	resp, err := client.ListAPIKeysWithResponse(ctx, spaceID)
	if err != nil {
		return nil, APIKeyListOutput{}, fmt.Errorf("failed to list API keys: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, APIKeyListOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}
	if resp.JSON200 == nil {
		return nil, APIKeyListOutput{}, errors.New("empty response from API")
	}

	return nil, APIKeyListOutput{APIKeys: *resp.JSON200}, nil
}
