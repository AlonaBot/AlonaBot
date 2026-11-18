package main

import (
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/spf13/cobra"
)

func checkCommand(code *int) *cobra.Command {
	var options generate.Options
	command := &cobra.Command{Use: "check [directory]", Short: "Check component resources and generated code without writing files", Args: cobra.MaximumNArgs(1)}
	componentFlags(command, &options)
	command.RunE = func(cmd *cobra.Command, args []string) error {
		options.Dir = componentDirectory(args)
		*code = exitFailure
		return pluginkit.Check(cmd.Context(), options)
	}
	return command
}
