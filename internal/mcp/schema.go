package mcp

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// SchemaInput represents input for ghost_schema
type SchemaInput struct {
	ID string `json:"id"`
}

func (SchemaInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[SchemaInput](nil))
	databaseRefInputProperties(schema)
	return schema
}

func newSchemaTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "ghost_schema",
		Title:       "Show Database Schema",
		Description: "Display database schema including tables, views, materialized views, and enum types with their columns, constraints, and indexes.",
		InputSchema: SchemaInput{}.Schema(),
		// No OutputSchema - we return raw text content
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Show Database Schema",
		},
	}
}

func (s *Server) handleSchema(ctx context.Context, req *mcp.CallToolRequest, input SchemaInput) (*mcp.CallToolResult, any, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, nil, err
	}

	schema, err := common.FetchDatabaseSchema(ctx, common.FetchDatabaseSchemaArgs{
		Client:      client,
		ProjectID:   projectID,
		DatabaseRef: input.ID,
	})
	if err != nil {
		return nil, nil, handleDatabaseError(err)
	}

	// Return raw text content
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: common.FormatSchema(schema)},
		},
	}, nil, nil
}
