package pluginkit

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/generate"
)

// Pack 构建单个平台的组件归档。
func Pack(ctx context.Context, options generate.Options, outDir string) (string, error) {
	archives, err := packTargets(ctx, options, []manifest.Platform{{OS: options.GOOS, Arch: options.GOARCH}}, outDir)
	if err != nil {
		return "", err
	}
	return archives[0], nil
}

// PackTargets 逐目标构建组件归档。
// targets 为空时使用本机平台；outDir 为空时写入工程 bin/。
func PackTargets(ctx context.Context, options generate.Options, targets []manifest.Platform, outDir string) (archives []string, err error) {
	if len(targets) == 0 {
		targets = []manifest.Platform{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
	}
	seen := map[manifest.Platform]bool{}
	for _, target := range targets {
		if !manifest.SupportedPlatform(target.OS, target.Arch) {
			return nil, fmt.Errorf("unsupported target %s/%s; supported targets: %v", target.OS, target.Arch, manifest.Platforms)
		}
		if seen[target] {
			return nil, fmt.Errorf("duplicate target %s/%s", target.OS, target.Arch)
		}
		seen[target] = true
	}
	return packTargets(ctx, options, targets, outDir)
}

func packTargets(ctx context.Context, options generate.Options, targets []manifest.Platform, outDir string) (archives []string, err error) {
	options.CGOEnabled = "0"
	options.RequireAdapterConstructor = true
	if outDir == "" {
		outDir = filepath.Join(options.Dir, "bin")
	}
	outDir, err = filepath.Abs(outDir)
	if err != nil {
		return nil, err
	}
	// 暂存目录与输出目录同级：无 /tmp 的容器也能打包，归档移动不跨文件系统。
	if err := os.MkdirAll(filepath.Dir(outDir), privateMode); err != nil {
		return nil, err
	}
	staging, err := os.MkdirTemp(filepath.Dir(outDir), ".alonabot-pack-*")
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(staging)) }()

	// 1. 每个平台分别扫描、构建与归档，全部成功后才写入输出目录。
	outputs := map[string][]byte{}
	for i, target := range targets {
		options.GOOS, options.GOARCH = target.OS, target.Arch
		snapshot, err := scanProject(ctx, options)
		if err != nil {
			return nil, err
		}
		name, compressed, err := packArchive(ctx, options, snapshot, filepath.Join(staging, fmt.Sprintf("build-%d", i)))
		if err != nil {
			return nil, err
		}
		if _, exists := outputs[name]; exists {
			continue
		} // 静态 WebUI 的 any 归档只发布一次。
		outputs[name] = compressed
		archives = append(archives, filepath.Join(outDir, name))
	}
	// 2. 临时文件全部写入成功后移动归档。
	if err := os.MkdirAll(outDir, privateMode); err != nil {
		return nil, err
	}
	for name, raw := range outputs {
		if err := os.WriteFile(filepath.Join(staging, name), raw, fileMode); err != nil {
			return nil, err
		}
	}
	for _, name := range slices.Sorted(maps.Keys(outputs)) {
		if err := os.Rename(filepath.Join(staging, name), filepath.Join(outDir, name)); err != nil {
			return nil, err
		}
	}
	return archives, nil
}

func packArchive(ctx context.Context, options generate.Options, snapshot *projectScan, buildDir string) (string, []byte, error) {
	files := snapshot.files
	// 归档资源取同一次构建的规范副本，薄入口与临时 module 信息不进入包。
	build, err := buildComponent(ctx, options, buildDir, false, snapshot)
	if err != nil {
		return "", nil, err
	}
	parsed, err := manifest.Parse(build.PackageManifest)
	if err != nil {
		return "", nil, err
	}
	m := parsed.Manifest
	for _, reserved := range []string{projectFile, "build-info", m.Entrypoint} {
		if _, exists := files[reserved]; exists {
			return "", nil, fmt.Errorf("resource %q conflicts with a generated package file", reserved)
		}
	}
	name := release.ArchiveName(m)
	if m.Entrypoint != "" {
		files[m.Entrypoint], err = os.ReadFile(build.Executable)
		if err != nil {
			return "", nil, err
		}
		info, err := buildinfo.ReadFile(build.Executable)
		if err != nil {
			return "", nil, fmt.Errorf("read executable build-info: %w", err)
		}
		files["build-info"] = []byte(info.String())
	}
	files[projectFile] = build.PackageManifest
	compressed, err := archiveFiles(files, m.Entrypoint)
	if err != nil {
		return "", nil, err
	}
	return name, compressed, nil
}

func archiveFiles(files map[string][]byte, executable string) ([]byte, error) {
	var output bytes.Buffer
	compressed := gzip.NewWriter(&output)
	archive := tar.NewWriter(compressed)
	seen := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(files)) {
		if err := manifest.ValidateRelPath(name); err != nil {
			return nil, err
		}
		for parent := name; parent != "."; parent = filepath.ToSlash(filepath.Dir(parent)) {
			if _, exists := files[parent]; exists && parent != name {
				return nil, fmt.Errorf("archive file %q conflicts with parent directory of %q", parent, name)
			}
			folded := strings.ToLower(parent)
			if previous, ok := seen[folded]; ok && previous != parent {
				return nil, fmt.Errorf("case-conflicting archive paths %q and %q", previous, parent)
			}
			seen[folded] = parent
		}
		mode := int64(0o644)
		if name == executable {
			mode = 0o755
		}
		if err := archive.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(files[name])), Typeflag: tar.TypeReg, Format: tar.FormatPAX}); err != nil {
			return nil, err
		}
		if _, err := archive.Write(files[name]); err != nil {
			return nil, err
		}
	}
	if err := errors.Join(archive.Close(), compressed.Close()); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
