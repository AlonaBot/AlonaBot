package main

import (
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/spf13/cobra"
)

func installMirrorFlags(command *cobra.Command) {
	command.Flags().String("github-api-base", "", "GitHub API base for this operation (empty: official API)")
	command.Flags().String("github-download-prefix", "", "Prefix prepended to the original GitHub asset URL for this operation (empty: direct download)")
}

func applyInstallMirrorFlags(command *cobra.Command, request *manage.InstallRequest) {
	if command.Flags().Changed("github-api-base") {
		value, _ := command.Flags().GetString("github-api-base")
		request.GitHubAPIBase = &value
	}
	if command.Flags().Changed("github-download-prefix") {
		value, _ := command.Flags().GetString("github-download-prefix")
		request.GitHubDownloadPrefix = &value
	}
}
