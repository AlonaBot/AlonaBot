package main

import (
	"github.com/AlonaBot/AlonaBot/internal/generate"
	"github.com/spf13/cobra"
)

func componentFlags(command *cobra.Command, options *generate.Options) {
	flags := command.Flags()
	flags.StringVar(&options.Dir, "dir", ".", "component project directory")
	flags.StringVar(&options.GOOS, "goos", "", "target operating system (default: Go environment)")
	flags.StringVar(&options.GOARCH, "goarch", "", "target architecture (default: Go environment)")
	flags.StringSliceVar(&options.Tags, "tags", nil, "comma-separated build tags")
}
