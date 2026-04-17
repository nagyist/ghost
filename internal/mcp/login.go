package mcp

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// LoginInput represents input for ghost_login
type LoginInput struct{}

func (LoginInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[LoginInput](nil))
}

// LoginOutput represents output for ghost_login
type LoginOutput struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	SpaceID string `json:"space_id"`
}

func (LoginOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[LoginOutput](nil))
	schema.Properties["name"].Description = "The authenticated user's name"
	schema.Properties["email"].Description = "The authenticated user's email"
	schema.Properties["space_id"].Description = "The authenticated user's space ID"
	return schema
}

func newLoginTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "ghost_login",
		Title: "Login",
		Description: `Authenticate with GitHub OAuth. Opens the user's browser to complete authentication.

Note: this tool requires user interaction in the browser. The user must complete the GitHub OAuth flow before the tool returns.`,
		InputSchema:  LoginInput{}.Schema(),
		OutputSchema: LoginOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(false),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Login",
		},
	}
}

func (s *Server) handleLogin(ctx context.Context, req *mcp.CallToolRequest, input LoginInput) (*mcp.CallToolResult, LoginOutput, error) {
	result, err := common.Login(ctx, s.app, false, nil)
	if err != nil {
		return nil, LoginOutput{}, err
	}

	return nil, LoginOutput{
		Name:    result.Name,
		Email:   result.Email,
		SpaceID: result.SpaceID,
	}, nil
}
