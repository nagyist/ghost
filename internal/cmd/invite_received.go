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

func buildInviteReceivedCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:   "received",
		Short: "List invitations you've received",
		Long: `List the pending invitations you've received across all spaces.

Accept one with 'ghost invite accept <space-id>' or decline it with
'ghost invite decline <space-id>'.`,
		Example: `  # List invitations you've received
  ghost invite received

  # Output as JSON
  ghost invite received --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, _, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListReceivedInvitesWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list invitations: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			invitations := *resp.JSON200

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), invitations)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), invitations)
			default:
				return outputReceivedInvites(cmd.OutOrStdout(), invitations)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputReceivedInvites(w io.Writer, invites []api.ReceivedInvite) error {
	table := common.NewTable(w)

	table.Header("SPACE ID", "SPACE", "FROM", "ROLE", "INVITED")
	for _, inv := range invites {
		from := inv.InviterEmail
		if from == "" {
			from = inv.InviterName
		}
		table.Append(inv.SpaceID, inv.SpaceName, from, string(inv.Role), inv.CreatedAt.Format(time.RFC3339))
	}
	return table.Render()
}
