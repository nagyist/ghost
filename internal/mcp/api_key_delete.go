package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// APIKeyDeleteInput represents input for ghost_api_key_delete.
type APIKeyDeleteInput struct {
	Prefix string `json:"prefix"`
}

func (APIKeyDeleteInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[APIKeyDeleteInput](nil))
	schema.Properties["prefix"].Description = "Prefix of the API key to delete (starts with 'gt_', as returned by ghost_api_key_list)"
	return schema
}

// APIKeyDeleteOutput represents output for ghost_api_key_delete.
type APIKeyDeleteOutput struct {
	Success bool   `json:"success"`
	Prefix  string `json:"prefix"`
}

func (APIKeyDeleteOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[APIKeyDeleteOutput](nil))
	successOutputProperties(schema)
	schema.Properties["prefix"].Description = "Prefix of the deleted API key"
	return schema
}

func newAPIKeyDeleteTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_api_key_delete",
		Title:        "Delete API Key",
		Description:  "Delete an API key from the current Ghost space by its prefix.",
		InputSchema:  APIKeyDeleteInput{}.Schema(),
		OutputSchema: APIKeyDeleteOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(true),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Delete API Key",
		},
	}
}

func (s *Server) handleAPIKeyDelete(ctx context.Context, req *mcp.CallToolRequest, input APIKeyDeleteInput) (*mcp.CallToolResult, APIKeyDeleteOutput, error) {
	cfg, client, spaceID, err := s.app.GetAll()
	if err != nil {
		return nil, APIKeyDeleteOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, APIKeyDeleteOutput{}, err
	}

	resp, err := client.DeleteAPIKeyWithResponse(ctx, spaceID, api.APIKeyPrefix(input.Prefix))
	if err != nil {
		return nil, APIKeyDeleteOutput{}, fmt.Errorf("failed to delete API key: %w", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		return nil, APIKeyDeleteOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	return nil, APIKeyDeleteOutput{
		Success: true,
		Prefix:  input.Prefix,
	}, nil
}
