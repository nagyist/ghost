package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildInviteSentCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:   "sent",
		Short: "List invites you've sent",
		Long:  `List the pending invites you've sent for the current space.`,
		Example: `  # List invites you've sent
  ghost invite sent

  # Output as JSON
  ghost invite sent --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListInvitesWithResponse(cmd.Context(), api.SpaceId(spaceID))
			if err != nil {
				return fmt.Errorf("failed to list invites: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			invites := *resp.JSON200

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), invites)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), invites)
			default:
				return outputInvites(cmd.OutOrStdout(), invites)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputInvites(w io.Writer, invites []api.Invite) error {
	table := common.NewTable(w)

	table.Header("EMAIL", "ROLE", "STATUS", "INVITED")
	for _, inv := range invites {
		table.Append(inv.Email, string(inv.Role), string(inv.Status), inv.CreatedAt.Format(time.RFC3339))
	}
	return table.Render()
}
