package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// InvoiceListInput represents input for ghost_invoice_list (no parameters).
type InvoiceListInput struct{}

func (InvoiceListInput) Schema() *jsonschema.Schema {
	return util.Must(jsonschema.For[InvoiceListInput](nil))
}

// InvoiceSummary represents a single invoice in the ghost_invoice_list output.
type InvoiceSummary struct {
	ID            string    `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceDate   time.Time `json:"invoice_date"`
	Total         float64   `json:"total"`
	Status        string    `json:"status"`
}

// InvoiceListOutput represents output for ghost_invoice_list.
type InvoiceListOutput struct {
	Invoices []InvoiceSummary `json:"invoices"`
}

func (InvoiceListOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[InvoiceListOutput](nil))
	inv := schema.Properties["invoices"].Items
	inv.Properties["id"].Description = "Opaque invoice identifier to pass to ghost_invoice for the line-item breakdown"
	inv.Properties["invoice_number"].Description = "Human-readable invoice number"
	inv.Properties["invoice_date"].Description = "Invoice date"
	inv.Properties["total"].Description = "Invoice total"
	inv.Properties["status"].Description = "Invoice status"
	inv.Properties["status"].Enum = []any{api.InvoiceStatusPaid, api.InvoiceStatusIssued, api.InvoiceStatusDelinquent}
	return schema
}

func newInvoiceListTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_invoice_list",
		Title:        "List Invoices",
		Description:  "List the most recent invoices.",
		InputSchema:  InvoiceListInput{}.Schema(),
		OutputSchema: InvoiceListOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:  true,
			OpenWorldHint: new(true),
			Title:         "List Invoices",
		},
	}
}

func (s *Server) handleInvoiceList(ctx context.Context, req *mcp.CallToolRequest, input InvoiceListInput) (*mcp.CallToolResult, InvoiceListOutput, error) {
	client, projectID, err := s.app.GetClient()
	if err != nil {
		return nil, InvoiceListOutput{}, err
	}

	resp, err := client.ListInvoicesWithResponse(ctx, projectID)
	if err != nil {
		return nil, InvoiceListOutput{}, fmt.Errorf("failed to list invoices: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, InvoiceListOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	if resp.JSON200 == nil {
		return nil, InvoiceListOutput{}, errors.New("empty response from API")
	}

	invoices := resp.JSON200.Invoices
	output := InvoiceListOutput{
		Invoices: make([]InvoiceSummary, len(invoices)),
	}
	for i, inv := range invoices {
		output.Invoices[i] = InvoiceSummary{
			ID:            inv.Id,
			InvoiceNumber: inv.InvoiceNumber,
			InvoiceDate:   inv.InvoiceDate,
			Total:         inv.Total,
			Status:        string(inv.Status),
		}
	}

	return nil, output, nil
}
