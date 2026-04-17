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

// DeleteInput represents input for ghost_delete
type DeleteInput struct {
	Ref string `json:"name_or_id"`
}

func (DeleteInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[DeleteInput](nil))
	databaseRefInputProperties(schema)
	return schema
}

// DeleteOutput represents output for ghost_delete
type DeleteOutput struct {
	Success     bool     `json:"success"`
	DeletedName string   `json:"deleted_name"`
	Warnings    []string `json:"warnings,omitempty"`
}

func (DeleteOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[DeleteOutput](nil))
	successOutputProperties(schema)
	schema.Properties["deleted_name"].Description = "Name of the deleted database"
	warningsOutputProperties(schema)
	return schema
}

func newDeleteTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_delete",
		Title:        "Delete Database",
		Description:  "Delete a database permanently. This operation is irreversible.",
		InputSchema:  DeleteInput{}.Schema(),
		OutputSchema: DeleteOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(true),
			IdempotentHint:  false,
			OpenWorldHint:   new(true),
			Title:           "Delete Database",
		},
	}
}

func (s *Server) handleDelete(ctx context.Context, req *mcp.CallToolRequest, input DeleteInput) (*mcp.CallToolResult, DeleteOutput, error) {
	cfg, client, projectID, err := s.app.GetAll()
	if err != nil {
		return nil, DeleteOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, DeleteOutput{}, err
	}

	// Fetch database details to get the name
	getResp, err := client.GetDatabaseWithResponse(ctx, projectID, input.Ref)
	if err != nil {
		return nil, DeleteOutput{}, fmt.Errorf("failed to get database details: %w", err)
	}

	if getResp.StatusCode() != http.StatusOK {
		return nil, DeleteOutput{}, common.ExitWithErrorFromStatusCode(getResp.StatusCode(), getResp.JSONDefault)
	}

	if getResp.JSON200 == nil {
		return nil, DeleteOutput{}, errors.New("empty response from API")
	}
	database := *getResp.JSON200
	databaseName := database.Name

	// Make the delete request using the resolved ID (no confirmation prompt for MCP - agents handle this)
	resp, err := client.DeleteDatabaseWithResponse(
		ctx,
		api.SpaceId(projectID),
		api.DatabaseRef(database.Id),
	)
	if err != nil {
		return nil, DeleteOutput{}, fmt.Errorf("failed to delete database: %w", err)
	}

	// Handle response
	if resp.StatusCode() != http.StatusAccepted {
		return nil, DeleteOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	// Remove the pgpass entry for the deleted database
	var warnings []string
	if err := common.RemovePgpassEntry(database, "tsdbadmin"); err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to remove .pgpass entry: %v", err))
	}

	return nil, DeleteOutput{
		Success:     true,
		DeletedName: databaseName,
		Warnings:    warnings,
	}, nil
}
