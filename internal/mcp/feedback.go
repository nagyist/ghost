package mcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
	"github.com/timescale/ghost/internal/util"
)

// FeedbackInput represents input for ghost_feedback
type FeedbackInput struct {
	Message string `json:"message"`
}

func (FeedbackInput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[FeedbackInput](nil))
	schema.Properties["message"].Description = "Feedback message, bug report, or support request"
	return schema
}

// FeedbackOutput represents output for ghost_feedback
type FeedbackOutput struct {
	Success bool `json:"success"`
}

func (FeedbackOutput) Schema() *jsonschema.Schema {
	schema := util.Must(jsonschema.For[FeedbackOutput](nil))
	successOutputProperties(schema)
	return schema
}

func newFeedbackTool() *mcp.Tool {
	return &mcp.Tool{
		Name:         "ghost_feedback",
		Title:        "Submit Feedback",
		Description:  "Submit feedback, a bug report, or a support request to the Ghost team.",
		InputSchema:  FeedbackInput{}.Schema(),
		OutputSchema: FeedbackOutput{}.Schema(),
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: new(false),
			IdempotentHint:  false,
			OpenWorldHint:   new(true),
			Title:           "Submit Feedback",
		},
	}
}

func (s *Server) handleFeedback(ctx context.Context, req *mcp.CallToolRequest, input FeedbackInput) (*mcp.CallToolResult, FeedbackOutput, error) {
	if input.Message == "" {
		return nil, FeedbackOutput{}, errors.New("feedback message cannot be empty")
	}

	client, _, err := s.app.GetClient()
	if err != nil {
		return nil, FeedbackOutput{}, err
	}

	resp, err := client.SubmitFeedbackWithResponse(ctx, api.SubmitFeedbackJSONRequestBody{
		Message: input.Message,
		Source:  "mcp",
		Version: config.Version,
		Os:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	})
	if err != nil {
		return nil, FeedbackOutput{}, fmt.Errorf("failed to submit feedback: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, FeedbackOutput{}, common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
	}

	return nil, FeedbackOutput{Success: true}, nil
}
