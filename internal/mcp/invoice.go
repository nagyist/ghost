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

// InvoiceInput represents input for ghost_invoice.
type InvoiceInput struct {
	InvoiceID string `json:"invoice_id"`
}

func (InvoiceInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[InvoiceInput](nil))
	schema.Properties["invoice_id"].Description = "Opaque invoice identifier returned by ghost_invoice_list"
	return schema
}

// InvoiceLineItem represents a single line item on an invoice.
type InvoiceLineItem struct {
	ProductType  string  `json:"product_type"`
	DatabaseID   string  `json:"database_id,omitempty"`
	DetailedSpec string  `json:"detailed_spec,omitempty"`
	Quantity     float64 `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	LineTotal    float64 `json:"line_total"`
}

// InvoiceOutput represents output for ghost_invoice.
type InvoiceOutput struct {
	LineItems []InvoiceLineItem `json:"line_items"`
}

func (InvoiceOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[InvoiceOutput](nil))
	li := schema.Properties["line_items"].Items
	li.Properties["product_type"].Description = "Product category"
	li.Properties["product_type"].Examples = []any{"storage", "compute"}
	li.Properties["database_id"].Description = "Ghost database ID this line item is attributed to, if any"
	li.Properties["detailed_spec"].Description = "Additional spec details (e.g. tier, size)"
	li.Properties["quantity"].Description = "Quantity billed"
	li.Properties["unit_price"].Description = "Unit price"
	li.Properties["line_total"].Description = "Line total"
	return schema
}

func newInvoiceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_invoice",
		Title:        "Get Invoice Detail",
		Description:  "Get the line-item breakdown for a single invoice, looked up by the opaque invoice ID returned from ghost_invoice_list.",
		InputSchema:  InvoiceInput{}.Schema(),
		OutputSchema: InvoiceOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "Get Invoice Detail",
		},
	}
}

func (s *Server) handleInvoice(ctx context.Context, req *mcp.CallToolRequest, input InvoiceInput) (*mcp.CallToolResult, InvoiceOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, InvoiceOutput{}, err
	}

	resp, err := client.GetInvoiceWithResponse(ctx, projectID, input.InvoiceID)
	if err != nil {
		return nil, InvoiceOutput{}, fmt.Errorf("failed to get invoice: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, InvoiceOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON200 == nil {
		return nil, InvoiceOutput{}, errors.New("empty response from API")
	}

	lineItems := resp.JSON200.LineItems
	output := InvoiceOutput{
		LineItems: make([]InvoiceLineItem, len(lineItems)),
	}
	for i, li := range lineItems {
		item := InvoiceLineItem{
			ProductType: li.ProductType,
			Quantity:    li.Quantity,
			UnitPrice:   li.UnitPrice,
			LineTotal:   li.LineTotal,
		}
		if li.DatabaseId != nil {
			item.DatabaseID = *li.DatabaseId
		}
		if li.DetailedSpec != nil {
			item.DetailedSpec = *li.DetailedSpec
		}
		output.LineItems[i] = item
	}

	return nil, output, nil
}
