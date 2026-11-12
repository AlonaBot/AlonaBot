package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
)

func checkPluginDependencies(pluginDir string) error {
	data, err := os.ReadFile(filepath.Join(pluginDir, "plugin.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read plugin.yaml: %w", err)
	}
	parsed, err := manifest.Parse(data)
	if err != nil {
		return err
	}
	// run 只启动当前开发插件，没有可满足依赖的其他已启用组件。
	installed := manifest.Installed{Plugins: map[string]string{}, Adapters: map[string][]manifest.ExtensionRequirement{}}
	var required []string
	for _, problem := range manifest.CheckDependencies(parsed.Manifest, installed) {
		if problem.Required {
			required = append(required, problem.Dependency+": "+problem.Reason)
		} else {
			slog.Warn("optional dependency is unavailable", "dependency", problem.Dependency, "reason", problem.Reason)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("required dependencies are unavailable: %s", strings.Join(required, "; "))
	}
	return nil
}
