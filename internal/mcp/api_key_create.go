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

// APIKeyCreateInput represents input for ghost_api_key_create.
type APIKeyCreateInput struct {
	Name string `json:"name,omitempty"`
}

func (APIKeyCreateInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[APIKeyCreateInput](nil))
	schema.Properties["name"].Description = "API key name (auto-generated if not provided)"
	return schema
}

// APIKeyCreateOutput represents output for ghost_api_key_create. APIKey is the
// one-time secret — it cannot be retrieved again later.
type APIKeyCreateOutput struct {
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
}

func (APIKeyCreateOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[APIKeyCreateOutput](nil))
	schema.Properties["name"].Description = "Name of the created API key"
	schema.Properties["api_key"].Description = "The API key secret. This is only returned once and cannot be retrieved again later — make sure to save it."
	return schema
}

func newAPIKeyCreateTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_api_key_create",
		Title:        "Create API Key",
		Description:  "Create a new API key for the current Ghost space. The secret is only returned once.",
		InputSchema:  APIKeyCreateInput{}.Schema(),
		OutputSchema: APIKeyCreateOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(false),
			IdempotentHint:  false,
			OpenWorldHint:   new(false),
			Title:           "Create API Key",
		},
	}
}

func (s *Server) handleAPIKeyCreate(ctx context.Context, req *mcp.CallToolRequest, input APIKeyCreateInput) (*mcp.CallToolResult, APIKeyCreateOutput, error) {
	cfg, client, spaceID, err := s.app.GetAll()
	if err != nil {
		return nil, APIKeyCreateOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, APIKeyCreateOutput{}, err
	}

	name := input.Name
	if name == "" {
		name, err = common.DefaultAPIKeyName(ctx, client)
		if err != nil {
			return nil, APIKeyCreateOutput{}, err
		}
	}

	resp, err := client.CreateAPIKeyWithResponse(ctx, spaceID, api.CreateAPIKeyJSONRequestBody{Name: name})
	if err != nil {
		return nil, APIKeyCreateOutput{}, fmt.Errorf("failed to create API key: %w", err)
	}
	if resp.StatusCode() != http.StatusCreated {
		return nil, APIKeyCreateOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}
	if resp.JSON201 == nil {
		return nil, APIKeyCreateOutput{}, errors.New("empty response from API")
	}

	return nil, APIKeyCreateOutput{
		Name:   name,
		APIKey: resp.JSON201.APIKey,
	}, nil
}
