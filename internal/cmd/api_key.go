package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildApiKeyCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-key",
		Short: "Manage API keys",
		Long: `Manage API keys for your Ghost space.

API keys can be used to authenticate with Ghost by setting the
GHOST_API_KEY environment variable.`,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
	}

	cmd.AddCommand(buildApiKeyCreateCmd(app))
	cmd.AddCommand(buildApiKeyListCmd(app))
	cmd.AddCommand(buildApiKeyDeleteCmd(app))

	return cmd
}
