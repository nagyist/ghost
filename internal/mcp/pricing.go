package mcp

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

type PricingInput struct{}

func (PricingInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[PricingInput](nil))
}

// PricingOutput is the MCP tool's output type. It has the same underlying type
// as common.PricingOutput so values convert directly, and is redeclared here so
// the tool can attach a Schema() method (matching the pattern other MCP tools
// use).
type PricingOutput common.PricingOutput

func (PricingOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[PricingOutput](nil))
	schema.Properties["dedicated"].Description = "Pricing for dedicated databases"

	ded := schema.Properties["dedicated"]
	ded.Properties["compute"].Description = "Per-size compute pricing, ordered smallest to largest"
	ded.Properties["storage"].Description = "Storage pricing"

	c := ded.Properties["compute"].Items
	c.Properties["size"].Description = "Size identifier"
	c.Properties["size"].Examples = []any{"1x", "2x", "4x", "8x"}
	c.Properties["milli_cpu"].Description = "CPU allocation in millicores (1000 = 1 vCPU)"
	c.Properties["memory_gib"].Description = "Memory allocation in GiB"
	c.Properties["price_per_hour"].Description = "Price per hour while the database is running"
	c.Properties["price_per_month"].Description = "Price per month while the database is running"

	st := ded.Properties["storage"]
	st.Properties["price_per_gib_hour"].Description = "Price per GiB per hour of provisioned storage above the included amount"
	st.Properties["price_per_gib_month"].Description = "Price per GiB per month of provisioned storage above the included amount"
	st.Properties["included_gib_per_database"].Description = "GiB of storage included per database at no additional charge. Only storage above this amount is billed at price_per_gib_hour"
	return schema
}

func newPricingTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_pricing",
		Title:        "Get Dedicated Pricing",
		Description:  "Get pricing for dedicated databases, including compute pricing for each available size and the storage rate.",
		InputSchema:  PricingInput{}.Schema(),
		OutputSchema: PricingOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Get Dedicated Pricing",
		},
	}
}

func (s *Server) handlePricing(ctx context.Context, req *mcp.CallToolRequest, input PricingInput) (*mcp.CallToolResult, PricingOutput, error) {
	client, _, err := s.app.GetClient()
	if err != nil {
		return nil, PricingOutput{}, err
	}

	output, err := common.FetchPricing(ctx, client)
	if err != nil {
		return nil, PricingOutput{}, err
	}
	return nil, PricingOutput(output), nil
}
