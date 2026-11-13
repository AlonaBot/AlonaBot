package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AlonaBot/AlonaBot"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
)

type developmentProject struct {
	dir     string
	build   pluginkit.Build
	sources map[string]string
	config  string
}

func runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options []alonabot.Option) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		if _, err := os.Stat("plugin.yaml"); err == nil {
			dirs = []string{"."}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	projects, err := buildDevelopment(ctx, dirs, dataDir)
	if err != nil {
		return err
	}
	for _, project := range projects {
		options = append(options, alonabot.Executable(project.build.Executable, project.build.PackageManifest))
	}
	if len(projects) > 0 {
		dataDir, err = filepath.EvalSymlinks(dataDir)
		if err != nil {
			return err
		}
	}
	if stdio {
		options = append(options, alonabot.Stdio())
	}
	core, err := alonabot.Start(ctx, options...)
	if err != nil {
		return err
	}
	var stopped chan struct{}
	if reload {
		stopped = make(chan struct{})
		go func() { defer close(stopped); pollDevelopment(ctx, core, projects, dataDir, interval) }()
	}
	err = core.Wait()
	cancel()
	if stopped != nil {
		<-stopped
	}
	return err
}

// buildDevelopment 在构建写入 dev/bin 前取得实例锁，避免覆盖正在运行的 Core 的产物。
func buildDevelopment(ctx context.Context, dirs []string, dataDir string) (projects []developmentProject, err error) {
	if len(dirs) == 0 {
		return nil, nil
	}
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	dataDir, err = filepath.EvalSymlinks(dataDir)
	if err != nil {
		return nil, err
	}
	projects = make([]developmentProject, 0, len(dirs))
	seen := make(map[string]bool)
	ids := make(map[string]bool)
	for _, dir := range dirs {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		dir, err = filepath.EvalSymlinks(dir)
		if err != nil {
			return nil, err
		}
		if seen[dir] {
			return nil, fmt.Errorf("duplicate development directory %s", dir)
		}
		seen[dir] = true
		raw, err := os.ReadFile(filepath.Join(dir, "plugin.yaml"))
		if err != nil {
			return nil, err
		}
		parsed, err := manifest.ParseProject(raw)
		if err != nil {
			return nil, err
		}
		id := parsed.Manifest.ID
		if ids[id] {
			return nil, fmt.Errorf("duplicate component ID %s", id)
		}
		ids[id] = true
		projects = append(projects, developmentProject{dir: dir})
	}
	for i := range projects {
		project := &projects[i]
		project.sources, err = developmentSources(project.dir, dataDir)
		if err != nil {
			return nil, err
		}
		project.build, err = pluginkit.BuildDev(ctx, project.dir, filepath.Join(dataDir, "dev", "bin"))
		if err != nil {
			return nil, err
		}
		project.config, err = developmentConfig(dataDir, project.build.ID)
		if err != nil {
			return nil, err
		}
	}
	return projects, nil
}

func developmentConfig(dataDir, id string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "plugins", id, "config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return "{}", nil
	}
	return string(raw), err
}

func pollDevelopment(ctx context.Context, core *alonabot.Core, projects []developmentProject, dataDir string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	instances, err := developmentInstances(dataDir)
	if err != nil {
		slog.Error("read development instances", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for i := range projects {
			project := &projects[i]
			sources, err := developmentSources(project.dir, dataDir)
			if err != nil {
				slog.Error("scan development project", "plugin_id", project.build.ID, "error", err)
				continue
			}
			configuration, err := developmentConfig(dataDir, project.build.ID)
			if err != nil {
				slog.Error("read development config", "plugin_id", project.build.ID, "error", err)
				continue
			}
			changed := !reflect.DeepEqual(sources, project.sources)
			if changed {
				project.sources = sources
				raw, err := os.ReadFile(filepath.Join(project.dir, "plugin.yaml"))
				if err == nil {
					var parsed *manifest.Result
					parsed, err = manifest.ParseProject(raw)
					if err == nil && (parsed.Manifest.ID != project.build.ID || parsed.Manifest.Role != project.build.Role || parsed.Manifest.AdapterID != project.build.AdapterID) {
						err = errors.New("development component identity changed")
					}
				}
				var build pluginkit.Build
				if err == nil {
					build, err = pluginkit.BuildDev(ctx, project.dir, filepath.Join(dataDir, "dev", "bin"))
				}
				if err != nil {
					slog.Error("development build failed", "plugin_id", project.build.ID, "error", err)
					continue
				}
				project.build = build
			}
			if changed || configuration != project.config {
				project.config = configuration
				if err := core.Restart(ctx, project.build.ID); err != nil {
					slog.Error("development restart failed", "plugin_id", project.build.ID, "error", err)
				}
			}
		}
		next, err := developmentInstances(dataDir)
		if err != nil {
			slog.Error("read development instances", "error", err)
			continue
		}
		for _, project := range projects {
			id := project.build.AdapterID
			if id == "" {
				continue
			}
			keys := make(map[string]bool)
			for instance := range instances[id] {
				keys[instance] = true
			}
			for instance := range next[id] {
				keys[instance] = true
			}
			for instance := range keys {
				if next[id][instance] != instances[id][instance] {
					if err := core.RestartInstance(ctx, id, instance); err != nil {
						slog.Error("development instance restart failed", "adapter_id", id, "instance_id", instance, "error", err)
					}
				}
			}
		}
		instances = next
	}
}

func developmentInstances(dataDir string) (map[string]map[string]string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "core_config.json"))
	if err != nil {
		return nil, err
	}
	var document struct {
		Adapters map[string]struct{ Instances map[string]json.RawMessage }
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	result := make(map[string]map[string]string)
	for id, group := range document.Adapters {
		result[id] = make(map[string]string)
		for instance, raw := range group.Instances {
			var value any
			decoder := json.NewDecoder(strings.NewReader(string(raw)))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				return nil, err
			}
			canonical, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			result[id][instance] = string(canonical)
		}
	}
	return result, nil
}

// developmentSources 按工程目录保存输入原文；比较内容可识别同大小、同时间戳的替换。
func developmentSources(dir, dataDir string) (map[string]string, error) {
	files := make(map[string]string)
	var embeds []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			if path != dir && (path == dataDir || name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = path
		if strings.HasSuffix(name, ".go") {
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, raw, parser.ParseComments)
			if err != nil {
				return nil
			} // 编译器在重建时给出语法错误原文。
			for _, group := range file.Comments {
				for _, comment := range group.List {
					patterns, ok := strings.CutPrefix(comment.Text, "//go:embed")
					if !ok || patterns == "" || (patterns[0] != ' ' && patterns[0] != '\t') {
						continue
					}
					for patterns != "" {
						patterns = strings.TrimSpace(patterns)
						if patterns == "" {
							break
						}
						var pattern string
						if patterns[0] == '"' || patterns[0] == '`' {
							end := 1
							for end < len(patterns) {
								if patterns[end] == patterns[0] && patterns[end-1] != '\\' {
									break
								}
								end++
							}
							if end == len(patterns) {
								break
							}
							pattern, err = strconv.Unquote(patterns[:end+1])
							if err != nil {
								return err
							}
							patterns = patterns[end+1:]
						} else {
							fields := strings.Fields(patterns)
							pattern = fields[0]
							patterns = patterns[len(pattern):]
						}
						pattern = strings.TrimPrefix(pattern, "all:")
						embeds = append(embeds, filepath.ToSlash(filepath.Join(filepath.Dir(relative), pattern)))
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for name := range files {
		keep := strings.HasSuffix(name, ".go") || name == "go.mod" || name == "go.sum" || name == "plugin.yaml" || name == "config_schema.json" || name == "adapter_config_schema.json" || strings.HasPrefix(name, "i18n/")
		for _, pattern := range embeds {
			for path := name; path != "."; path = filepath.ToSlash(filepath.Dir(path)) {
				matched, err := filepath.Match(pattern, path)
				if err != nil {
					return nil, err
				}
				keep = keep || matched
			}
		}
		if !keep {
			delete(files, name)
			continue
		}
		raw, err := os.ReadFile(files[name])
		if err != nil {
			return nil, err
		}
		files[name] = string(raw)
	}
	return files, nil
}
