package mcp

import (
	"context"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// StatusInput represents input for ghost_status (empty - no parameters)
type StatusInput struct{}

func (StatusInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[StatusInput](nil))
}

// StatusOutput represents output for ghost_status
type StatusOutput struct {
	ComputeMinutes      int64                 `json:"compute_minutes"`
	ComputeLimitMinutes int64                 `json:"compute_limit_minutes"`
	Storage             string                `json:"storage"`
	StorageLimit        string                `json:"storage_limit"`
	Databases           common.DatabaseCounts `json:"databases"`
	CostToDate          *float64              `json:"cost_to_date,omitempty"`
	EstimatedTotalCost  *float64              `json:"estimated_total_cost,omitempty"`
	BillingPeriodStart  *time.Time            `json:"billing_period_start,omitempty"`
	BillingPeriodEnd    *time.Time            `json:"billing_period_end,omitempty"`
}

func (StatusOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[StatusOutput](nil))
	schema.Properties["compute_minutes"].Description = "Current compute usage in minutes"
	schema.Properties["compute_limit_minutes"].Description = "Compute limit in minutes"
	schema.Properties["storage"].Description = "Current storage usage"
	schema.Properties["storage"].Examples = []any{"512MiB", "1GiB"}
	schema.Properties["storage_limit"].Description = "Storage limit"
	schema.Properties["storage_limit"].Examples = []any{"8GiB", "16GiB"}
	schema.Properties["databases"].Description = "Number of databases in each status"
	schema.Properties["cost_to_date"].Description = "Net cost accrued so far this billing cycle"
	schema.Properties["estimated_total_cost"].Description = "Projected net total for the current billing cycle based on usage to date"
	schema.Properties["billing_period_start"].Description = "Start of the current billing cycle"
	schema.Properties["billing_period_end"].Description = "End of the current billing cycle"
	return schema
}

func newStatusTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_status",
		Title:        "Show Space Usage",
		Description:  "Display database space usage including compute minutes, storage, and database counts by status.",
		InputSchema:  StatusInput{}.Schema(),
		OutputSchema: StatusOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Show Space Usage",
		},
	}
}

func (s *Server) handleStatus(ctx context.Context, req *mcp.CallToolRequest, input StatusInput) (*mcp.CallToolResult, StatusOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, StatusOutput{}, err
	}

	status, err := common.FetchStatus(ctx, client, projectID)
	if err != nil {
		return nil, StatusOutput{}, err
	}

	return nil, StatusOutput{
		ComputeMinutes:      status.ComputeMinutes,
		ComputeLimitMinutes: status.ComputeLimitMinutes,
		Storage:             common.FormatStorageSize(new(int(status.StorageMib))),
		StorageLimit:        common.FormatStorageSize(new(int(status.StorageLimitMib))),
		Databases:           status.Databases,
		CostToDate:          status.CostToDate,
		EstimatedTotalCost:  status.EstimatedTotalCost,
		BillingPeriodStart:  status.BillingPeriodStart,
		BillingPeriodEnd:    status.BillingPeriodEnd,
	}, nil
}
