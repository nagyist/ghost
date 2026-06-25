package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildSpaceCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:   "space",
		Short: "Show or manage spaces",
		Long: `Show or manage Ghost spaces.

A space is a collection of databases with shared usage limits and billing.
The CLI operates on one space at a time — the current space. Running
'ghost space' with no subcommand shows details about the current space.`,
		Example: `  # Show the current space
  ghost space

  # Output as JSON
  ghost space --json

  # Output as YAML
  ghost space --yaml`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.GetSpaceWithResponse(cmd.Context(), spaceID)
			if err != nil {
				return fmt.Errorf("failed to get space: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}
			space := resp.JSON200

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), space)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), space)
			default:
				outputSpace(cmd, space)
				return nil
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	cmd.AddCommand(buildSpaceListCmd(app))
	cmd.AddCommand(buildSpaceUseCmd(app))
	cmd.AddCommand(buildSpaceRenameCmd(app))

	return cmd
}

func outputSpace(cmd *cobra.Command, space *api.SpaceDetail) {
	cmd.Printf("Space: %s (%s)\n", space.Name, space.Id)
	if space.Owner != nil {
		cmd.Printf("Owner: %s (%s)\n", space.Owner.Name, space.Owner.Email)
	}
	if space.Role != nil {
		cmd.Printf("Role: %s\n", *space.Role)
	}
}
