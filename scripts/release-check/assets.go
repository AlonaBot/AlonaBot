package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
)

type component struct {
	id, role, source string
}

// 整合清单与 scripts/dist 相同，Adapter 与 help 由各自仓库按需安装。
var components = []component{
	{"alonabot-agent-official", "agent", "github.com/AlonaBot/alonabot-agent"},
	{"alonabot-webui-admin", "webui", "github.com/AlonaBot/alonabot-webui"},
}

func verifyAssets(opt options, r *report) error {
	if err := r.perform("installer", func() error {
		raw, err := readRegular(filepath.Join(opt.candidate, "install.sh"))
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			return errors.New("install.sh is empty")
		}
		return nil
	}); err != nil {
		return err
	}
	publications := map[string]*manifest.Manifest{}
	if opt.bundle {
		if err := r.perform("package-inventory", func() error {
			entries, err := os.ReadDir(filepath.Join(opt.candidate, "packages"))
			if err != nil {
				return err
			}
			var actual, expected []string
			for _, entry := range entries {
				if !entry.IsDir() {
					return fmt.Errorf("package entry %q is not a directory", entry.Name())
				}
				actual = append(actual, entry.Name())
			}
			for _, c := range components {
				expected = append(expected, c.id)
			}
			return sameFiles(actual, expected)
		}); err != nil {
			return err
		}
		// 1. 每个平台只收录组件归档，身份与资源直接从包内读取。
		for _, c := range components {
			index := len(r.Components)
			r.Components = append(r.Components, componentResult{ID: c.id, Source: c.source, Assets: []assetResult{}})
			var expectedFiles []string
			for _, target := range r.Platforms {
				goos, goarch, _ := strings.Cut(target, "/")
				if err := r.perform("package/"+c.id+"/"+target, func() error {
					directory := filepath.Join(opt.candidate, "packages", c.id)
					entries, err := os.ReadDir(directory)
					if err != nil {
						return err
					}
					publication := release.GitHubRelease{}
					for _, entry := range entries {
						if entry.Type().IsRegular() {
							publication.Assets = append(publication.Assets, release.Download{Name: entry.Name()})
						}
					}
					asset, err := publication.Select(goos, goarch)
					if err != nil {
						return err
					}
					filename := filepath.Join(directory, asset.Name)
					info, err := os.Stat(filename)
					if err != nil {
						return err
					}
					destination := filepath.Join(opt.out, c.id+"-"+goos+"-"+goarch)
					files, err := extractArchive(filename, destination)
					if err != nil {
						return err
					}
					raw, err := readRegular(filepath.Join(destination, "plugin.yaml"))
					if err != nil {
						return err
					}
					parsed, err := manifest.Parse(raw)
					if err != nil {
						return err
					}
					m := parsed.Manifest
					if release.ArchiveName(m) != asset.Name {
						return errors.New("archive name differs from package manifest")
					}
					if previous, ok := publications[c.id]; ok && previous.Version != m.Version {
						return errors.New("component platform versions differ")
					}
					resources, err := checkPackage(destination, files, c, m.Version, goos, goarch)
					if err != nil {
						return err
					}
					publications[c.id] = m
					r.Components[index].Version = m.Version
					r.Components[index].Assets = append(r.Components[index].Assets, assetResult{target, filename, info.Size(), resources})
					expectedFiles = append(expectedFiles, asset.Name)
					return nil
				}); err != nil {
					return err
				}
			}
			if err := r.perform("inventory/"+c.id, func() error {
				entries, err := os.ReadDir(filepath.Join(opt.candidate, "packages", c.id))
				if err != nil {
					return err
				}
				var actual []string
				for _, entry := range entries {
					actual = append(actual, entry.Name())
				}
				return sameFiles(actual, expectedFiles)
			}); err != nil {
				return err
			}
		}
	}
	// 2. 核对归档摘要与 Core 身份；整合模式另外核对组件副本。
	var sums map[string]string
	if err := r.perform("SHA256SUMS", func() error {
		var err error
		sums, err = readSums(filepath.Join(opt.candidate, "SHA256SUMS"))
		if err == nil && len(sums) != len(r.Platforms) {
			err = fmt.Errorf("SHA256SUMS has %d archives; want %d", len(sums), len(r.Platforms))
		}
		return err
	}); err != nil {
		return err
	}
	for _, target := range r.Platforms {
		if err := r.perform("distribution/"+target, func() error {
			goos, goarch, _ := strings.Cut(target, "/")
			rootName := fmt.Sprintf("alonabot-%s-%s-%s", opt.version, goos, goarch)
			extension := ".tar.gz"
			if goos == "windows" {
				extension = ".zip"
			}
			name := rootName + extension
			filename := filepath.Join(opt.candidate, name)
			size, digest, err := archiveDigest(filename)
			if err != nil {
				return err
			}
			if sums[name] != digest {
				return fmt.Errorf("%s SHA256 = %s; SHA256SUMS declares %q", name, digest, sums[name])
			}
			destination := filepath.Join(opt.out, "distribution-"+goos+"-"+goarch)
			files, err := extractArchive(filepath.Join(opt.candidate, name), destination)
			if err != nil {
				return err
			}
			coreName := "alonabot"
			if goos == "windows" {
				coreName += ".exe"
			}
			core := filepath.Join(destination, rootName, coreName)
			if err := checkExecutable(core, goos); err != nil {
				return err
			}
			if _, err := checkBuildInfo(core, goos, goarch); err != nil {
				return err
			}
			values, err := linkedStrings(core, goos, versionSymbol)
			if err != nil {
				return err
			}
			if string(values[versionSymbol]) != opt.version {
				return fmt.Errorf("Core linked version differs from %q", opt.version)
			}
			for _, name := range []string{"LICENSE", "source.json"} {
				raw, err := readRegular(filepath.Join(destination, rootName, name))
				if err != nil {
					return err
				}
				if len(raw) == 0 {
					return fmt.Errorf("empty distribution resource %s", name)
				}
			}
			sourceRaw, err := readRegular(filepath.Join(destination, rootName, "source.json"))
			if err != nil {
				return err
			}
			var source struct {
				Source     string                     `json:"source"`
				Version    string                     `json:"version"`
				Components map[string]json.RawMessage `json:"components"`
			}
			if err := json.Unmarshal(sourceRaw, &source); err != nil {
				return err
			}
			if source.Source != "https://github.com/AlonaBot/AlonaBot/tree/v"+opt.version || source.Version != opt.version {
				return errors.New("distribution source/version differs from release")
			}
			if !opt.bundle && len(source.Components) != 0 {
				return errors.New("Core archive declares bundled components")
			}
			expected := []string{rootName + "/" + coreName, rootName + "/LICENSE", rootName + "/source.json"}
			if opt.bundle {
				for _, c := range components {
					m := *publications[c.id]
					m.OS, m.Arch = goos, goarch
					for _, file := range []string{release.ArchiveName(&m)} {
						rel := filepath.Join("packages", c.id, file)
						outside, err := readRegular(filepath.Join(opt.candidate, rel))
						if err != nil {
							return err
						}
						inside, err := readRegular(filepath.Join(destination, rootName, rel))
						if err != nil {
							return err
						}
						if !bytes.Equal(outside, inside) {
							return fmt.Errorf("inner and external %s differ", rel)
						}
						expected = append(expected, rootName+"/"+filepath.ToSlash(rel))
					}
				}
			}
			if err := sameFiles(files, expected); err != nil {
				return err
			}
			r.Distributions = append(r.Distributions, distributionResult{target, filename, size, digest})
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func archiveDigest(filename string) (int64, string, error) {
	info, err := os.Lstat(filename)
	if err != nil {
		return 0, "", err
	}
	if !info.Mode().IsRegular() {
		return 0, "", fmt.Errorf("%s is not a regular file", filename)
	}
	file, err := os.Open(filename)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	return size, hex.EncodeToString(hash.Sum(nil)), err
}

func readSums(filename string) (map[string]string, error) {
	raw, err := readRegular(filename)
	if err != nil {
		return nil, err
	}
	sums := map[string]string{}
	for _, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		digest, name, ok := strings.Cut(line, "  ")
		decoded, err := hex.DecodeString(digest)
		if !ok || err != nil || len(decoded) != sha256.Size || strings.ToLower(digest) != digest || name != filepath.Base(name) || name == "." || name == "" {
			return nil, fmt.Errorf("invalid SHA256SUMS line %q", line)
		}
		if _, exists := sums[name]; exists {
			return nil, fmt.Errorf("duplicate SHA256SUMS archive %q", name)
		}
		sums[name] = digest
	}
	return sums, nil
}

func checkPackage(dir string, files []string, c component, version, goos, goarch string) ([]string, error) {
	entrypoint := "component"
	if goos == "windows" {
		entrypoint += ".exe"
	}
	manifest, err := readRegular(filepath.Join(dir, "plugin.yaml"))
	if err != nil {
		return nil, err
	}
	// dist 使用 pack 的稳定 YAML 输出；这里只核对其固定标量记录，完整语义由正式安装入口检查。
	identityRecords := []string{"id: " + c.id, "version: " + version, "source: " + c.source, "role: " + c.role}
	expected := append(slices.Clone(identityRecords), "os: "+goos, "arch: "+goarch, "entrypoint: "+entrypoint, "manifest_version: 1")
	lines := strings.Split(string(manifest), "\n")
	for _, record := range expected {
		field, _, _ := strings.Cut(record, ":")
		count := 0
		for _, line := range lines {
			if strings.HasPrefix(line, field+":") {
				if line != record {
					return nil, fmt.Errorf("package manifest record %q; want %q", line, record)
				}
				count++
			}
		}
		if count != 1 {
			return nil, fmt.Errorf("package manifest must contain exactly one canonical record %q; got %d", record, count)
		}
	}
	executable := filepath.Join(dir, entrypoint)
	if err := checkExecutable(executable, goos); err != nil {
		return nil, err
	}
	info, err := checkBuildInfo(executable, goos, goarch)
	if err != nil {
		return nil, err
	}
	raw, err := readRegular(filepath.Join(dir, "build-info"))
	if err != nil {
		return nil, err
	}
	if _, err := debug.ParseBuildInfo(string(raw)); err != nil {
		return nil, fmt.Errorf("parse build-info: %w", err)
	}
	if !bytes.Equal(raw, []byte(info.String())) {
		return nil, errors.New("build-info differs from the entrypoint")
	}
	binary, err := readRegular(executable)
	if err != nil {
		return nil, err
	}
	// 生成器内嵌作者 manifest；包内 manifest 另外增加平台、入口与协议字段。
	// 因而只对共有的身份记录作静态比对，运行时资源登记由正式加载验收。
	for _, record := range identityRecords {
		if err := checkEmbedded(binary, []byte(record+"\n"), "author manifest record "+record); err != nil {
			return nil, err
		}
	}
	resources := []string{"plugin.yaml:author-identity-records"}
	for _, name := range files {
		if name != "config_schema.json" && name != "adapter_config_schema.json" && !(filepath.ToSlash(filepath.Dir(name)) == "i18n" && strings.HasSuffix(name, ".json")) {
			continue
		}
		raw, err := readRegular(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			return nil, err
		}
		if !json.Valid(raw) {
			return nil, fmt.Errorf("invalid JSON resource %s", name)
		}
		if err := checkEmbedded(binary, raw, name); err != nil {
			return nil, err
		}
		resources = append(resources, name)
	}
	slices.Sort(resources)
	return resources, nil
}

func checkExecutable(filename, goos string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || goos != "windows" && info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("missing executable permissions on %s", filename)
	}
	return nil
}
