package cmd

import (
	"net/http"
	"strings"
	"testing"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/api/mock"
)

func TestMCPListCmd(t *testing.T) {
	wantText := "TYPE    NAME                                    \n" +
		"prompt  design-postgis-tables                   \n" +
		"prompt  design-postgres-tables                  \n" +
		"prompt  find-hypertable-candidates              \n" +
		"prompt  migrate-postgres-tables-to-hypertables  \n" +
		"prompt  pgvector-semantic-search                \n" +
		"prompt  postgres                                \n" +
		"prompt  postgres-database-migration             \n" +
		"prompt  postgres-hybrid-text-search             \n" +
		"prompt  setup-timescaledb-hypertables           \n" +
		"tool    ghost_api_key_create                    \n" +
		"tool    ghost_api_key_delete                    \n" +
		"tool    ghost_api_key_list                      \n" +
		"tool    ghost_connect                           \n" +
		"tool    ghost_create                            \n" +
		"tool    ghost_create_dedicated                  \n" +
		"tool    ghost_delete                            \n" +
		"tool    ghost_feedback                          \n" +
		"tool    ghost_fork                              \n" +
		"tool    ghost_fork_dedicated                    \n" +
		"tool    ghost_id                                \n" +
		"tool    ghost_invoice                           \n" +
		"tool    ghost_invoice_list                      \n" +
		"tool    ghost_list                              \n" +
		"tool    ghost_login                             \n" +
		"tool    ghost_logs                              \n" +
		"tool    ghost_mcp_tool_refresh                  \n" +
		"tool    ghost_password                          \n" +
		"tool    ghost_pause                             \n" +
		"tool    ghost_pricing                           \n" +
		"tool    ghost_rename                            \n" +
		"tool    ghost_resume                            \n" +
		"tool    ghost_schema                            \n" +
		"tool    ghost_share                             \n" +
		"tool    ghost_share_list                        \n" +
		"tool    ghost_share_revoke                      \n" +
		"tool    ghost_sql                               \n" +
		"tool    ghost_usage                             \n" +
		"tool    search_docs                             \n" +
		"tool    view_skill                              \n"

	// With function_tools disabled, the ghost_mcp_tool_refresh management tool
	// is not registered, so it drops out of the listing.
	wantTextNoRefresh := strings.Replace(wantText,
		"tool    ghost_mcp_tool_refresh                  \n", "", 1)

	tests := []cmdTest{
		{
			name:       "text output",
			args:       []string{"mcp", "list"},
			wantStdout: wantText,
		},
		{
			name:       "ls alias",
			args:       []string{"mcp", "ls"},
			wantStdout: wantText,
		},
		{
			// JSON output is 1000+ lines (full tool schemas), so we just verify
			// it doesn't error. The text test above validates the capability list.
			name: "json output",
			args: []string{"mcp", "list", "--json"},
		},
		{
			// Same rationale as JSON.
			name: "yaml output",
			args: []string{"mcp", "list", "--yaml"},
		},
		{
			// --function-tools connects to and lists every database in the
			// space; with none defined, the tool list is unchanged from the
			// default (ghost_mcp_tool_refresh is already registered either
			// way).
			name: "function-tools flag with no databases",
			args: []string{"mcp", "list", "--function-tools"},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-space").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &[]api.DatabaseWithUsage{},
					}, nil)
			},
			wantStdout: wantText,
		},
		{
			// function_tools=false disables the feature entirely, so the
			// ghost_mcp_tool_refresh management tool is not registered.
			name:       "function_tools disabled",
			args:       []string{"mcp", "list"},
			opts:       []runOption{withEnv("GHOST_FUNCTION_TOOLS", "false")},
			wantStdout: wantTextNoRefresh,
		},
		{
			// The explicit --function-tools flag overrides function_tools=false:
			// it still connects to every database and registers the management
			// tool (flags take precedence over config).
			name: "function-tools flag overrides disabled config",
			args: []string{"mcp", "list", "--function-tools"},
			opts: []runOption{withEnv("GHOST_FUNCTION_TOOLS", "false")},
			setup: func(m *mock.MockClientWithResponsesInterface) {
				m.EXPECT().ListDatabasesWithResponse(validCtx, "test-space").
					Return(&api.ListDatabasesResponse{
						HTTPResponse: httpResponse(http.StatusOK),
						JSON200:      &[]api.DatabaseWithUsage{},
					}, nil)
			},
			wantStdout: wantText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCommand(t, tt.args, tt.setup, tt.opts...)

			if tt.wantErr != "" {
				if result.err == nil {
					t.Fatal("expected error, got nil")
				}
				assertOutput(t, result.err.Error(), tt.wantErr)
			} else if result.err != nil {
				t.Fatalf("unexpected error: %v", result.err)
			}

			if tt.wantStdout != "" {
				assertOutput(t, result.stdout, tt.wantStdout)
			} else if result.stdout == "" {
				t.Error("expected non-empty stdout")
			}
		})
	}
}
