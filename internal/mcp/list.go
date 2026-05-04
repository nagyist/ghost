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

// ListInput represents input for ghost_list (empty - no parameters)
type ListInput struct{}

func (ListInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[ListInput](nil))
}

// DatabaseInfo represents information about a single database
type DatabaseInfo struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Type           api.DatabaseType   `json:"type"`
	Size           *api.DatabaseSize  `json:"size,omitempty"`
	Status         api.DatabaseStatus `json:"status"`
	Storage        string             `json:"storage"`
	ComputeMinutes *int64             `json:"compute_minutes,omitempty"`
}

// ListOutput represents output for ghost_list
type ListOutput struct {
	Databases []DatabaseInfo `json:"databases"`
}

func (ListOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[ListOutput](nil))
	dbInfo := schema.Properties["databases"].Items
	databaseIDOutputProperties(dbInfo)
	databaseNameOutputProperties(dbInfo)
	dbInfo.Properties["type"].Description = "Database type"
	dbInfo.Properties["size"].Description = "Compute size for dedicated databases"
	dbInfo.Properties["status"].Description = "Database status"
	dbInfo.Properties["storage"].Description = "Current storage usage"
	dbInfo.Properties["storage"].Examples = []any{"512MiB", "1GiB"}
	dbInfo.Properties["compute_minutes"].Description = "Compute minutes used by this database during the current billing cycle. Only populated for standard databases."
	return schema
}

func newListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_list",
		Title:        "List Databases",
		Description:  "List all databases, including each database's current status, storage usage, and compute minutes used in the current billing cycle.",
		InputSchema:  ListInput{}.Schema(),
		OutputSchema: ListOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "List Databases",
		},
	}
}

func (s *Server) handleList(ctx context.Context, req *mcp.CallToolRequest, input ListInput) (*mcp.CallToolResult, ListOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, ListOutput{}, err
	}

	// Make API call to list databases
	resp, err := client.ListDatabasesWithResponse(ctx, projectID)
	if err != nil {
		return nil, ListOutput{}, fmt.Errorf("failed to list databases: %w", err)
	}

	// Handle API response
	if resp.StatusCode() != http.StatusOK {
		return nil, ListOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON200 == nil {
		return nil, ListOutput{}, errors.New("empty response from API")
	}
	databases := *resp.JSON200

	// Build output
	output := make([]DatabaseInfo, len(databases))
	for i, database := range databases {
		output[i] = DatabaseInfo{
			ID:             database.Id,
			Name:           database.Name,
			Type:           database.Type,
			Size:           database.Size,
			Status:         database.Status,
			Storage:        common.FormatStorageSize(database.StorageMib),
			ComputeMinutes: database.ComputeMinutes,
		}
	}

	return nil, ListOutput{
		Databases: output,
	}, nil
}
