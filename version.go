package alonabot

import "runtime/debug"

const coreModule = "github.com/AlonaBot/AlonaBot"

func coreVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "(devel)"
	}
	return moduleVersion(info)
}

func moduleVersion(info *debug.BuildInfo) string {
	if info.Main.Path == coreModule && info.Main.Version != "" {
		return info.Main.Version
	}
	for _, dependency := range info.Deps {
		if dependency.Path == coreModule && dependency.Version != "" {
			return dependency.Version
		}
	}
	return "(devel)"
}
