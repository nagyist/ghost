package cmd

import (
	"errors"
	"testing"
)

func TestLogoutCmd(t *testing.T) {
	tests := []cmdTest{
		{
			name:    "not logged in",
			args:    []string{"logout"},
			opts:    []runOption{withClientError(errors.New("authentication required: no credentials found"))},
			wantErr: "authentication required: no credentials found",
		},
	}

	runCmdTests(t, tests)
}
