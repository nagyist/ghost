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

// PauseInput represents input for ghost_pause
type PauseInput struct {
	Ref string `json:"name_or_id"`
}

func (PauseInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[PauseInput](nil))
	databaseRefInputProperties(schema)
	return schema
}

// PauseOutput represents output for ghost_pause
type PauseOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func (PauseOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[PauseOutput](nil))
	successOutputProperties(schema)
	databaseIDOutputProperties(schema)
	databaseNameOutputProperties(schema)
	return schema
}

func newPauseTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_pause",
		Title:        "Pause Database",
		Description:  "Pause a running database. This terminates active connections.",
		InputSchema:  PauseInput{}.Schema(),
		OutputSchema: PauseOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(true),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Pause Database",
		},
	}
}

func (s *Server) handlePause(ctx context.Context, req *mcp.CallToolRequest, input PauseInput) (*mcp.CallToolResult, PauseOutput, error) {
	cfg, client, projectID, err := s.app.GetAll()
	if err != nil {
		return nil, PauseOutput{}, err
	}

	if err := checkReadOnly(cfg); err != nil {
		return nil, PauseOutput{}, err
	}

	// Make the pause request
	resp, err := client.PauseDatabaseWithResponse(
		ctx,
		api.SpaceId(projectID),
		api.DatabaseRef(input.Ref),
	)
	if err != nil {
		return nil, PauseOutput{}, fmt.Errorf("failed to pause database: %w", err)
	}

	// Handle API response
	if resp.StatusCode() != http.StatusAccepted {
		return nil, PauseOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON202 == nil {
		return nil, PauseOutput{}, errors.New("empty response from API")
	}
	database := *resp.JSON202

	return nil, PauseOutput{
		Success: true,
		ID:      database.Id,
		Name:    database.Name,
	}, nil
}
