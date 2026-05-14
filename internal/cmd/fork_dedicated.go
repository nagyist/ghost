package cmd

import (
	"github.com/spf13/cobra"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/common"
	"github.com/timescale/ghost/internal/util"
)

func buildForkDedicatedCmd(app *common.App) *cobra.Command {
	var name string
	var size string
	var jsonOutput bool
	var yamlOutput bool
	var wait bool

	cmd := &cobra.Command{
		Use:   "dedicated <name-or-id>",
		Short: "Fork a database as dedicated",
		Long: `Fork an existing database as a new dedicated instance. The fork inherits
the source database's data but runs as an always-on, billed instance.
A payment method must be on file.

Run 'ghost pricing' to see compute and storage pricing.`,
		Example: `  # Fork as dedicated with default size (1x)
  ghost fork dedicated my-database

  # Fork with a specific size
  ghost fork dedicated my-database --size 4x

  # Fork with a custom name
  ghost fork dedicated my-database --name myapp-dedicated

  # Fork and output as JSON
  ghost fork dedicated my-database --json

  # Fork and wait for the database to be ready
  ghost fork dedicated my-database --size 2x --wait`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: databaseCompletion(app),
		SilenceUsage:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return forkDatabase(cmd, app, forkDatabaseArgs{
				sourceDatabaseRef: args[0],
				req: api.ForkDatabaseRequest{
					Name: util.PtrIfNonZero(name),
					Type: new(api.DatabaseTypeDedicated),
					Size: new(api.DatabaseSize(size)),
				},
				jsonOutput: jsonOutput,
				yamlOutput: yamlOutput,
				wait:       wait,
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the forked database (auto-generated if not provided)")
	cmd.Flags().StringVar(&size, "size", "1x", "Database size (1x, 2x, 4x, 8x)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	cmd.Flags().BoolVar(&yamlOutput, "yaml", false, "Output in YAML format")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for the database to be ready before returning")
	cmd.MarkFlagsMutuallyExclusive("json", "yaml")

	if err := cmd.RegisterFlagCompletionFunc("size", sizeCompletion); err != nil {
		panic(err)
	}

	return cmd
}
