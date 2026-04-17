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

// ResumeInput represents input for ghost_resume
type ResumeInput struct {
	ID   string `json:"id"`
	Wait bool   `json:"wait,omitempty"`
}

func (ResumeInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[ResumeInput](nil))
	databaseRefInputProperties(schema)
	waitInputProperties(schema)
	return schema
}

// ResumeOutput represents output for ghost_resume
type ResumeOutput struct {
	ID               string   `json:"id"`
	ConnectionString string   `json:"connection_string"`
	Warnings         []string `json:"warnings,omitempty"`
}

func (ResumeOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[ResumeOutput](nil))
	databaseIDOutputProperties(schema)
	connectionStringOutputProperties(schema)
	warningsOutputProperties(schema)
	return schema
}

func newResumeTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_resume",
		Title:        "Resume Database",
		Description:  "Resume a paused database.",
		InputSchema:  ResumeInput{}.Schema(),
		OutputSchema: ResumeOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(false),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
			Title:           "Resume Database",
		},
	}
}

func (s *Server) handleResume(ctx context.Context, req *mcp.CallToolRequest, input ResumeInput) (*mcp.CallToolResult, ResumeOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, ResumeOutput{}, err
	}

	// Make the start request
	resp, err := client.ResumeDatabaseWithResponse(
		ctx,
		api.SpaceId(projectID),
		api.DatabaseRef(input.ID),
	)
	if err != nil {
		return nil, ResumeOutput{}, fmt.Errorf("failed to resume database: %w", err)
	}

	// Handle API response
	if resp.StatusCode() != http.StatusAccepted {
		return nil, ResumeOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON202 == nil {
		return nil, ResumeOutput{}, errors.New("empty response from API")
	}

	database := *resp.JSON202

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
		return nil, ResumeOutput{}, fmt.Errorf("failed to build connection string: %w", err)
	}

	if input.Wait {
		if err := common.WaitForDatabase(ctx, common.WaitForDatabaseArgs{
			Client:      client,
			ProjectID:   projectID,
			DatabaseRef: database.Id,
		}); err != nil {
			return nil, ResumeOutput{}, err
		}
	}

	return nil, ResumeOutput{
		ID:               database.Id,
		ConnectionString: connStr,
		Warnings:         warnings,
	}, nil
}
