package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildConfigSetCmd(app *common.App) *cobra.Command {
	return &cobra.Command{
		Use:               "set <key> <value>",
		Short:             "Set configuration value",
		Long:              `Set a configuration value and save it to ~/.config/ghost/config.yaml`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: configOptionCompletion,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]

			cfg := app.GetConfig()
			if err := cfg.Set(key, value); err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}

			cmd.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}
}
