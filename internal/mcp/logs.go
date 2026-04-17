package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// LogsInput represents input for ghost_logs
type LogsInput struct {
	Ref   string    `json:"name_or_id"`
	Tail  int       `json:"tail,omitempty"`
	Until time.Time `json:"until,omitempty"`
}

func (LogsInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[LogsInput](nil))
	databaseRefInputProperties(schema)
	schema.Properties["tail"].Description = "Number of log lines to fetch"
	schema.Properties["tail"].Default = util.Must(json.Marshal(500))
	schema.Properties["tail"].Minimum = new(1.0)
	schema.Properties["until"].Description = "Fetch logs before this timestamp (RFC3339 format, e.g., 2024-01-15T10:00:00Z). If not provided, fetches up to the current time."
	return schema
}

// LogsOutput represents output for ghost_logs
type LogsOutput struct {
	Logs []string `json:"logs"`
}

func (LogsOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[LogsOutput](nil))
	schema.Properties["logs"].Description = "Array of log entries, ordered from oldest to newest"
	return schema
}

func newLogsTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_logs",
		Title:        "Fetch Database Logs",
		Description:  "Fetch logs from a database. Returns log entries in chronological order (oldest first). By default, returns the last 500 log entries.",
		InputSchema:  LogsInput{}.Schema(),
		OutputSchema: LogsOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Fetch Database Logs",
		},
	}
}

func (s *Server) handleLogs(ctx context.Context, req *mcp.CallToolRequest, input LogsInput) (*mcp.CallToolResult, LogsOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, LogsOutput{}, err
	}

	logs, err := common.FetchLogs(ctx, common.FetchLogsArgs{
		Client:      client,
		ProjectID:   projectID,
		DatabaseRef: input.Ref,
		Tail:        input.Tail,
		Until:       input.Until,
	})
	if err != nil {
		return nil, LogsOutput{}, err
	}

	return nil, LogsOutput{Logs: logs}, nil
}
