package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
	"github.com/timescale/ghost/internal/config"
)

func feedbackRequest(message string) api.SubmitFeedbackJSONRequestBody {
	return api.SubmitFeedbackJSONRequestBody{
		Message: message,
		Source:  "cli",
		Version: config.Version,
		Os:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func TestFeedbackCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"feedback", "hello"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
		{
			name:    "empty message from stdin",
			args:    []string{"feedback"},
			opts:    []runOption{withStdin("")},
			wantErr: "feedback message cannot be empty",
		},
		{
			name: "network error",
			args: []string{"feedback", "test"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().SubmitFeedbackWithResponse(validCtx, feedbackRequest("test")).
					Return(nil, errors.New("connection refused"))
			},
			wantErr: "failed to submit feedback: connection refused",
		},
		{
			name: "API error",
			args: []string{"feedback", "test"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().SubmitFeedbackWithResponse(validCtx, feedbackRequest("test")).
					Return(&api.SubmitFeedbackResponse{
						HTTPResponse: httpResponse(http.StatusInternalServerError),
						JSONDefault:  &api.Error{Message: "internal error"},
					}, nil)
			},
			wantErr: "internal error",
		},
		{
			name: "message from argument",
			args: []string{"feedback", "Great tool!"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().SubmitFeedbackWithResponse(validCtx, feedbackRequest("Great tool!")).
					Return(&api.SubmitFeedbackResponse{
						HTTPResponse: httpResponse(http.StatusOK),
					}, nil)
			},
			wantStdout: "Feedback submitted! Thank you.\n",
		},
		{
			name: "message from stdin",
			args: []string{"feedback"},
			opts: []runOption{withStdin("Great tool!")},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().SubmitFeedbackWithResponse(validCtx, feedbackRequest("Great tool!")).
					Return(&api.SubmitFeedbackResponse{
						HTTPResponse: httpResponse(http.StatusOK),
					}, nil)
			},
			wantStdout: "Feedback submitted! Thank you.\n",
		},
	}

	runCmdTests(t, tests)
}
