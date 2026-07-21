package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildMemberListCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List space members",
		Long:    `List the members of the current space.`,
		Example: `  # List the members of the current space
  ghost member list

  # Output as JSON
  ghost member list --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListMembersWithResponse(cmd.Context(), api.SpaceID(spaceID))
			if err != nil {
				return fmt.Errorf("failed to list members: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			members := *resp.JSON200

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), members)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), members)
			default:
				return outputMembers(cmd.OutOrStdout(), members)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputMembers(w io.Writer, members []api.Member) error {
	table := common.NewTable(w)

	table.Header("NAME", "EMAIL", "ROLE")
	for _, m := range members {
		table.Append(m.Name, m.Email, string(m.Role))
	}
	return table.Render()
}
