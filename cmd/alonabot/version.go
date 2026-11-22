package main

import (
	"strings"

	"github.com/spf13/cobra"
)

// 发布构建通过 -ldflags 的 -X main.version 和 -X main.buildDate 注入版本与日期。
var (
	version   = "dev"
	buildDate string
)

func configureVersion(cmd *cobra.Command) {
	cmd.Version = strings.TrimPrefix(version, "v")
	format := "{{.Name}} version {{.Version}}"
	if buildDate != "" {
		format += " (" + buildDate + ")"
	}
	format += "\nhttps://github.com/AlonaBot/AlonaBot"
	if cmd.Version != "dev" {
		format += "/releases/tag/v{{.Version}}"
	}
	cmd.SetVersionTemplate(format + "\n")
}
