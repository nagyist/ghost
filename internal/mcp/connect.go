package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// ConnectInput represents input for ghost_connect
type ConnectInput struct {
	Ref string `json:"name_or_id"`
}

func (ConnectInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[ConnectInput](nil))
	databaseRefInputProperties(schema)
	return schema
}

// ConnectOutput represents output for ghost_connect
type ConnectOutput struct {
	ConnectionString string   `json:"connection_string"`
	Warnings         []string `json:"warnings,omitempty"`
}

func (ConnectOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[ConnectOutput](nil))
	connectionStringOutputProperties(schema)
	warningsOutputProperties(schema)
	return schema
}

func newConnectTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_connect",
		Title:        "Get Connection String",
		Description:  "Get a PostgreSQL connection string for a database.",
		InputSchema:  ConnectInput{}.Schema(),
		OutputSchema: ConnectOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Get Connection String",
		},
	}
}

func (s *Server) handleConnect(ctx context.Context, req *mcp.CallToolRequest, input ConnectInput) (*mcp.CallToolResult, ConnectOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, ConnectOutput{}, err
	}

	// Fetch database details
	resp, err := client.GetDatabaseWithResponse(ctx, projectID, input.Ref)
	if err != nil {
		return nil, ConnectOutput{}, fmt.Errorf("failed to get database: %w", err)
	}

	// Handle API response
	if resp.StatusCode() != http.StatusOK {
		return nil, ConnectOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON200 == nil {
		return nil, ConnectOutput{}, errors.New("empty response from API")
	}

	database := *resp.JSON200

	// Get password for database
	var warnings []string
	password, err := common.GetPassword(database, "tsdbadmin")
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to retrieve password: %v", err))
	}

	// Build connection string
	connStr, err := common.BuildConnectionString(common.ConnectionStringArgs{
		Database: database,
		Role:     "tsdbadmin",
		Password: password,
	})
	if err != nil {
		return nil, ConnectOutput{}, fmt.Errorf("failed to build connection string: %w", err)
	}

	return nil, ConnectOutput{
		ConnectionString: connStr,
		Warnings:         warnings,
	}, nil
}
