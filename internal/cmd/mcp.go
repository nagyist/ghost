package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
	"github.com/timescale/ghost/internal/mcp"
)

// buildMCPCmd creates the MCP server command with subcommands
func buildMCPCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Ghost Model Context Protocol (MCP) server",
		Long: `Ghost Model Context Protocol (MCP) server for AI assistant integration.

Exposes Ghost CLI functionality as MCP tools for Claude and other AI assistants.`,
	}

	// Add subcommands
	cmd.AddCommand(buildMCPInstallCmd(app))
	cmd.AddCommand(buildMCPUninstallCmd(app))
	cmd.AddCommand(buildMCPStatusCmd(app))
	cmd.AddCommand(buildMCPStartCmd(app))
	cmd.AddCommand(buildMCPListCmd(app))
	cmd.AddCommand(buildMCPGetCmd(app))

	return cmd
}

// addFunctionToolsFlag registers the --function-tools flag, shared by `mcp
// list` and `mcp get`: it additionally includes each database's generated
// (@mcp) function tools in the listing, by connecting to and introspecting
// every database in the space rather than just registering the built-in
// ghost_* tools and the refresh management tool.
func addFunctionToolsFlag(cmd *cobra.Command, functionTools *bool) {
	cmd.Flags().BoolVar(functionTools, "function-tools", false,
		"Also include each database's generated custom function tools (connects to every database in the space)")
}

// functionToolsMode picks the function-tool mode for `mcp list`/`mcp get`.
//
// The explicit --function-tools flag (functionTools) always enables the full
// feature — connecting to and introspecting every database — overriding the
// function_tools config option, since it's an explicit request that has no
// other purpose (flags take precedence over config). Otherwise the config
// decides: when function_tools is disabled the feature is off entirely; when
// enabled, only the refresh management tool is registered, so listings stay
// accurate without connecting to any database.
func functionToolsMode(cfg *config.Config, functionTools bool) mcp.FunctionToolsMode {
	switch {
	case functionTools:
		return mcp.FunctionToolsEnabled
	case cfg.FunctionTools:
		return mcp.FunctionToolsManagementOnly
	default:
		return mcp.FunctionToolsDisabled
	}
}
