package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io"
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

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/app"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/pluginkit"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

type developmentProject struct {
	dir     string
	assets  string // WebUI 直接读取的静态根目录，无构建或进程。
	build   pluginkit.Build
	sources map[string]string
	config  string
}

func consoleEnabled(dataDir string, explicit, enabled, terminal bool) (bool, error) {
	if explicit {
		return enabled, nil
	}
	if !terminal {
		return false, nil
	}
	raw, err := os.ReadFile(filepath.Join(dataDir, "core_config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	var configuration struct {
		Adapters map[string]struct {
			Instances map[string]json.RawMessage `json:"instances"`
		} `json:"adapters"`
	}
	if err := json.Unmarshal(raw, &configuration); err != nil {
		return false, fmt.Errorf("read console adapter settings: %w", err)
	}
	for _, group := range configuration.Adapters {
		if len(group.Instances) != 0 {
			return false, nil
		}
	}
	return true, nil
}

func runDevelopment(dirs []string, dataDir string, reload, stdio bool, interval time.Duration, options app.Options) error {
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
		if project.assets != "" {
			options.WebUIs = append(options.WebUIs, os.DirFS(project.assets))
		} else {
			options.Components = append(options.Components, app.ExecutableComponent(project.build.Executable, project.build.PackageManifest))
		}
	}
	if len(projects) > 0 {
		dataDir, err = filepath.EvalSymlinks(dataDir)
		if err != nil {
			return err
		}
	}
	options.Stdio = stdio
	var console *consoleAdapter
	if _, enabled := options.RuntimeAdapters[consoleAdapterID]; enabled {
		console = &consoleAdapter{input: os.Stdin, output: os.Stdout, terminal: isTerminal(os.Stdout), inputTerminal: isTerminal(os.Stdin)}
		options.Components = append(options.Components, app.AdapterComponent(console))
	}
	core, err := app.Start(ctx, options)
	if err == context.Canceled {
		return nil
	}
	if err != nil {
		return err
	}
	if console != nil {
		go func() {
			if err := console.readInput(); err != nil {
				console.mu.Lock()
				instance := console.current
				console.mu.Unlock()
				slog.Error("console input failed", "error", err)
				if reportErr := instance.core.ReportFailure(ctx, errcode.Internal.New(err.Error())); reportErr != nil {
					slog.Error("report console failure", "error", reportErr)
				}
			}
		}()
	}
	var stopped chan struct{}
	if reload && len(projects) != 0 {
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
	webUI := false
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
		project := developmentProject{dir: dir}
		if parsed.Manifest.Role == "webui" {
			if webUI {
				return nil, errors.New("only one development WebUI directory is allowed")
			}
			webUI = true
			project.assets = filepath.Join(dir, filepath.FromSlash(parsed.Manifest.Assets))
		}
		projects = append(projects, project)
	}
	for i := range projects {
		project := &projects[i]
		if project.assets != "" {
			continue
		}
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

func pollDevelopment(ctx context.Context, core *app.Core, projects []developmentProject, dataDir string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	settings, err := developmentSettings(dataDir, projects)
	if err != nil {
		slog.Error("read development settings", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for i := range projects {
			project := &projects[i]
			if project.assets != "" {
				continue
			}
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
		next, err := developmentSettings(dataDir, projects)
		if err != nil {
			slog.Error("read development settings", "error", err)
			continue
		}
		// Core 设置先发布，实例在同一轮重启时即可读取新的管理员、前缀与语言。
		if next.core != settings.core {
			if err := core.ReloadSettings(ctx); err != nil {
				slog.Error("development settings reload failed", "error", err)
				continue
			}
		}
		for _, project := range projects {
			id := project.build.AdapterID
			if id == "" {
				continue
			}
			keys := make(map[string]bool)
			for instance := range settings.instances[id] {
				keys[instance] = true
			}
			for instance := range next.instances[id] {
				keys[instance] = true
			}
			for instance := range keys {
				if next.instances[id][instance] != settings.instances[id][instance] {
					if err := core.RestartInstance(ctx, id, instance); err != nil {
						slog.Error("development instance restart failed", "adapter_id", id, "instance_id", instance, "error", err)
					}
				}
			}
		}
		settings = next
	}
}

type developmentSettingsSnapshot struct {
	instances map[string]map[string]string
	core      string // 排除开发 Adapter 实例后的规范 JSON，其他设置变化只触发 ReloadSettings。
}

func developmentSettings(dataDir string, projects []developmentProject) (developmentSettingsSnapshot, error) {
	result := developmentSettingsSnapshot{instances: make(map[string]map[string]string)}
	raw, err := os.ReadFile(filepath.Join(dataDir, "core_config.json"))
	if err != nil {
		return result, err
	}
	var document map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return result, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			err = errors.New("unexpected value after core settings")
		}
		return result, err
	}
	adapters := make(map[string]any)
	if value, exists := document["adapters"]; exists {
		var ok bool
		adapters, ok = value.(map[string]any)
		if !ok {
			return result, errors.New("adapters must be an object")
		}
	}
	for _, project := range projects {
		id := project.build.AdapterID
		if id == "" {
			continue
		}
		value, exists := adapters[id]
		if !exists {
			continue
		}
		group, ok := value.(map[string]any)
		if !ok {
			return result, fmt.Errorf("adapters.%s must be an object", id)
		}
		value, exists = group["instances"]
		if !exists {
			continue
		}
		instances, ok := value.(map[string]any)
		if !ok {
			return result, fmt.Errorf("adapters.%s.instances must be an object", id)
		}
		result.instances[id] = make(map[string]string)
		for instance, value := range instances {
			canonical, err := json.Marshal(value)
			if err != nil {
				return result, err
			}
			result.instances[id][instance] = string(canonical)
		}
		delete(group, "instances")
	}
	canonical, err := json.Marshal(document)
	result.core = string(canonical)
	return result, err
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
