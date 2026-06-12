package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

// Space represents a space in CLI output.
type Space struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Current bool   `json:"current"`
}

func buildSpaceListCmd(app *common.App) *cobra.Command {
	var jsonOutput bool
	var yamlOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List spaces",
		Long:    `List your Ghost spaces. The current space is marked with an asterisk.`,
		Example: `  # List your spaces
  ghost space list

  # Output as JSON
  ghost space list --json`,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, spaceID, err := app.GetClient()
			if err != nil {
				return err
			}

			resp, err := client.ListSpacesWithResponse(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to list spaces: %w", err)
			}
			if resp.StatusCode() != http.StatusOK {
				return common.ExitWithErrorFromStatusCode(resp.StatusCode(), resp.JSONDefault)
			}
			if resp.JSON200 == nil {
				return errors.New("empty response from API")
			}

			spaces := *resp.JSON200
			output := make([]Space, len(spaces))
			for i, s := range spaces {
				output[i] = Space{
					ID:      s.Id,
					Name:    s.Name,
					Current: s.Id == spaceID,
				}
			}

			switch {
			case jsonOutput:
				return util.SerializeToJSON(cmd.OutOrStdout(), output)
			case yamlOutput:
				return util.SerializeToYAML(cmd.OutOrStdout(), output)
			default:
				return outputSpaces(cmd.OutOrStdout(), output)
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	return cmd
}

func outputSpaces(w io.Writer, spaces []Space) error {
	table := common.NewTable(w)

	table.Header("ID", "NAME")
	for _, s := range spaces {
		name := s.Name
		if s.Current {
			name += " *"
		}
		table.Append(s.ID, name)
	}
	return table.Render()
}
