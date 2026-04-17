package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildConfigUnsetCmd(app *common.App) *cobra.Command {
	return &cobra.Command{
		Use:               "unset <key>",
		Short:             "Remove configuration value",
		Long:              `Remove a configuration value and save changes to ~/.config/ghost/config.yaml`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: configOptionCompletion,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			cfg := app.GetConfig()
			if err := cfg.Unset(key); err != nil {
				return fmt.Errorf("failed to unset config: %w", err)
			}

			cmd.Printf("Unset %s\n", key)
			return nil
		},
	}
}
