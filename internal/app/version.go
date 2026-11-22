package app

import "runtime/debug"

const coreModule = "github.com/AlonaBot/AlonaBot"

// releaseVersion 由发行构建经 -ldflags -X 注入；开发构建为空，改从模块构建信息读取。
var releaseVersion string

// Version 返回发行构建注入的版本，开发构建改从模块构建信息读取。
func Version() string {
	if releaseVersion != "" {
		return releaseVersion
	}
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
		if dependency.Path == coreModule {
			if dependency.Replace != nil {
				dependency = dependency.Replace
			}
			if dependency.Version != "" {
				return dependency.Version
			}
			return "(devel)"
		}
	}
	return "(devel)"
}
