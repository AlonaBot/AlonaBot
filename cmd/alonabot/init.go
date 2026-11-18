package main

import (
	"fmt"

	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/spf13/cobra"
)

func initCommand(code *int) *cobra.Command {
	var id, role, adapterID, module string
	command := &cobra.Command{Use: "init [directory]", Short: "Create a component project in an empty directory", Args: cobra.MaximumNArgs(1)}
	flags := command.Flags()
	flags.StringVar(&id, "id", "", "component ID (default: directory name)")
	flags.StringVar(&role, "role", "plugin", "component role: plugin, adapter, agent or webui")
	flags.StringVar(&adapterID, "adapter-id", "", "platform ID (default: component ID for adapters)")
	flags.StringVar(&module, "module", "", "Go module path (default: directory name outside an existing module)")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		*code = exitFailure
		dir := componentDirectory(args)
		needsSDK, err := pluginkit.Init(cmd.Context(), dir, id, role, adapterID, module)
		if err != nil {
			return err
		}
		if needsSDK {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Development CLI: run go get github.com/AlonaBot/AlonaBot@<version> to select an SDK version."); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created component project. Change source in plugin.yaml before publishing.")
		return err
	}
	return command
}
