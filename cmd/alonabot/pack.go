package main

import (
	"fmt"

	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/spf13/cobra"
)

func packCommand(code *int) *cobra.Command {
	var options generate.Options
	var assetURL, key, out string
	command := &cobra.Command{Use: "pack", Short: "Build a component archive and optionally sign release.json", Args: cobra.NoArgs}
	componentFlags(command, &options)
	flags := command.Flags()
	flags.StringVar(&assetURL, "asset-url", "", "absolute https asset URL (required)")
	flags.StringVar(&key, "key", "", "standard base64 Ed25519 private key file (64 bytes)")
	flags.StringVar(&out, "out", "", "output directory (default: <project>/bin)")
	command.MarkFlagRequired("asset-url")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		*code = exitFailure
		path, err := pluginkit.Pack(cmd.Context(), options, assetURL, key, out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Packed", path)
		return err
	}
	return command
}
