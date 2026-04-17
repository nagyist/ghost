package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/config"
	"github.com/timescale/ghost/internal/util"
)

func buildFeedbackCmd(app *common.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback [message]",
		Short: "Submit feedback, a bug report, or a support request",
		Long: `Submit feedback, a bug report, or a support request to the Ghost team.

If no message is provided as an argument, reads from stdin.`,
		Example: `  # Submit feedback as an argument
  ghost feedback "I can't connect to my database after resuming it"

  # Submit feedback from stdin
  echo "Great tool!" | ghost feedback

  # Submit feedback interactively
  ghost feedback
  # → Enter your feedback (press Ctrl+D when done):`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := app.GetClient()
			if err != nil {
				return err
			}

			// Get message from args or stdin
			var message string
			if len(args) > 0 {
				message = args[0]
			} else {
				if util.IsTerminal(cmd.InOrStdin()) {
					cmd.PrintErrln("Enter your feedback (press Ctrl+D when done):")
				}
				input, err := util.ReadAll(cmd.Context(), cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("failed to read feedback: %w", err)
				}
				message = input
			}

			if message == "" {
				return errors.New("feedback message cannot be empty")
			}

			resp, err := client.SubmitFeedbackWithResponse(cmd.Context(), api.SubmitFeedbackJSONRequestBody{
				Message: message,
				Source:  "cli",
				Version: config.Version,
				Os:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			})
			if err != nil {
				return fmt.Errorf("failed to submit feedback: %w", err)
			}

			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}

			cmd.Println("Feedback submitted! Thank you.")
			return nil
		},
	}

	return cmd
}
