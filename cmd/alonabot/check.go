package main

import (
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/sdk/generate"
	"github.com/spf13/cobra"
)

func checkCommand(code *int) *cobra.Command {
	var options generate.Options
	command := &cobra.Command{Use: "check", Short: "Check component resources and generated code without writing files", Args: cobra.NoArgs}
	componentFlags(command, &options)
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		*code = exitFailure
		return pluginkit.Check(cmd.Context(), options)
	}
	return command
}
