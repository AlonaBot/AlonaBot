package main

import (
	"fmt"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/spf13/cobra"
)

func packCommand(code *int) *cobra.Command {
	var options generate.Options
	var out string
	var targets []string
	command := &cobra.Command{Use: "pack [directory]", Short: "Build component archives for target platforms", Args: cobra.MaximumNArgs(1)}
	flags := command.Flags()
	flags.StringArrayVar(&targets, "target", nil, "target OS/arch (repeatable; default: host platform)")
	flags.StringSliceVar(&options.Tags, "tags", nil, "comma-separated build tags")
	flags.StringVar(&out, "out", "", "output directory (default: <project>/bin)")
	command.RunE = func(cmd *cobra.Command, args []string) error {
		*code = exitFailure
		options.Dir = componentDirectory(args)
		var platforms []manifest.Platform
		for _, target := range targets {
			goos, goarch, ok := strings.Cut(target, "/")
			if !ok || !manifest.SupportedPlatform(goos, goarch) {
				return fmt.Errorf("unsupported target %q; supported targets: %v", target, manifest.Platforms)
			}
			platforms = append(platforms, manifest.Platform{OS: goos, Arch: goarch})
		}
		paths, err := pluginkit.PackTargets(cmd.Context(), options, platforms, out)
		if err != nil {
			return err
		}
		for _, path := range paths {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Packed", path); err != nil {
				return err
			}
		}
		return nil
	}
	return command
}
