package main

import (
	"fmt"

	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/sdk/generate"
	"github.com/spf13/cobra"
)

func buildCommand(code *int) *cobra.Command {
	var options generate.Options
	var out string
	command := &cobra.Command{Use: "build", Short: "Generate resources and build an independent component", Args: cobra.NoArgs}
	componentFlags(command, &options)
	command.Flags().StringVar(&out, "out", "", "output directory (default: <project>/bin)")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		*code = exitFailure
		build, err := pluginkit.Compile(cmd.Context(), options, out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Built %s (%s): %s\n", build.ID, build.Role, build.Executable)
		return err
	}
	return command
}
