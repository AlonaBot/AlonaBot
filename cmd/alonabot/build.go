package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/spf13/cobra"
)

func buildCommand(code *int) *cobra.Command {
	var options generate.Options
	var out string
	command := &cobra.Command{Use: "build", Short: "Build a component or library application", Args: cobra.NoArgs}
	componentFlags(command, &options)
	command.Flags().StringVar(&out, "out", "", "output directory (default: <project>/bin)")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		*code = exitFailure
		if _, err := os.Stat(filepath.Join(options.Dir, "plugin.yaml")); errors.Is(err, os.ErrNotExist) {
			executable, err := pluginkit.CompileApp(cmd.Context(), options, out)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Built application: %s\n", executable)
			return err
		} else if err != nil {
			return err
		}
		build, err := pluginkit.Compile(cmd.Context(), options, out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Built %s (%s): %s\n", build.ID, build.Role, build.Executable)
		return err
	}
	return command
}
