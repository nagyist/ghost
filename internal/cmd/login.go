package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
)

func buildLoginCmd(app *common.App) *cobra.Command {
	var headless bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub OAuth",
		Long: `Authenticate via GitHub OAuth. Opens your browser to complete authentication.

Use --headless for environments without a browser (Docker containers, SSH
sessions, CI/CD, etc.).`,
		Example: `  # Login via browser (default)
  ghost login

  # Login from a headless environment (Docker, SSH, CI/CD)
  ghost login --headless`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := common.Login(cmd.Context(), app, headless, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			cmd.Printf("Successfully logged in as %s\n", result.Email)

			return nil
		},
	}

	cmd.Flags().BoolVar(&headless, "headless", false,
		"Use device authorization flow (for environments without a browser)")

	return cmd
}
