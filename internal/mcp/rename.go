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

// RenameInput represents input for ghost_rename
type RenameInput struct {
	Ref  string `json:"name_or_id"`
	Name string `json:"name"`
}

func (RenameInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[RenameInput](nil))
	databaseRefInputProperties(schema)
	schema.Properties["name"].Description = "New name for the database"
	return schema
}

// RenameOutput represents output for ghost_rename
type RenameOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	OldName string `json:"old_name"`
	Name    string `json:"name"`
}

func (RenameOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[RenameOutput](nil))
	successOutputProperties(schema)
	databaseIDOutputProperties(schema)
	schema.Properties["old_name"].Description = "Previous name of the database"
	schema.Properties["name"].Description = "New name of the database"
	return schema
}

func newRenameTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_rename",
		Title:        "Rename Database",
		Description:  "Rename a database.",
		InputSchema:  RenameInput{}.Schema(),
		OutputSchema: RenameOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(false),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Rename Database",
		},
	}
}

func (s *Server) handleRename(ctx context.Context, req *mcp.CallToolRequest, input RenameInput) (*mcp.CallToolResult, RenameOutput, error) {
	cfg, client, projectID, err := s.app.GetAll()
	if err != nil {
		return nil, RenameOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, RenameOutput{}, err
	}

	// Fetch database details to resolve name/ID
	getResp, err := client.GetDatabaseWithResponse(ctx, projectID, input.Ref)
	if err != nil {
		return nil, RenameOutput{}, fmt.Errorf("failed to get database details: %w", err)
	}

	if getResp.StatusCode() != http.StatusOK {
		return nil, RenameOutput{}, common.ExitWithErrorFromStatusCode(getResp.StatusCode(), getResp.JSONDefault)
	}

	if getResp.JSON200 == nil {
		return nil, RenameOutput{}, errors.New("empty response from API")
	}
	database := *getResp.JSON200

	resp, err := client.RenameDatabaseWithResponse(
		ctx,
		api.SpaceId(projectID),
		api.DatabaseRef(database.Id),
		api.RenameDatabaseRequest{Name: input.Name},
	)
	if err != nil {
		return nil, RenameOutput{}, fmt.Errorf("failed to rename database: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return nil, RenameOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	return nil, RenameOutput{
		Success: true,
		ID:      database.Id,
		OldName: database.Name,
		Name:    input.Name,
	}, nil
}
