package diagnostics

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"
)

const (
	formatVersion = 1
	dateLayout    = "2006-01-02"
)

var logName = regexp.MustCompile(`^alonabot-(\d{4}-\d{2}-\d{2})(?:\.[1-9]\d*)?\.log$`)

// Options 选择诊断内容；Since 必须为正，三组内容至少选择一组。
type Options struct {
	Since                              time.Duration
	Logs, Settings, Components, Redact bool
}

type manifest struct {
	FormatVersion int       `json:"format_version"`
	GeneratedAt   time.Time `json:"generated_at"`
	Since         string    `json:"since"`
	From          time.Time `json:"from"`
	To            time.Time `json:"to"`
	Logs          bool      `json:"logs"`
	Settings      bool      `json:"settings"`
	Components    bool      `json:"components"`
	Redact        bool      `json:"redact"`
	Files         []string  `json:"files"`
}

// Write 收集所选内容并写出 zip；调用方负责在错误时丢弃输出。
func Write(ctx context.Context, w io.Writer, src Source, opts Options) (err error) {
	if opts.Since <= 0 {
		return errors.New("since must be a positive duration")
	}
	if !opts.Logs && !opts.Settings && !opts.Components {
		return errors.New("select at least one diagnostic group")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// 1. 固定时间范围，读取来源；调用方在错误时丢弃归档。
	now := time.Now().UTC()
	snapshot, err := src.load(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, snapshot.close()) }()
	archive := zip.NewWriter(w)
	defer func() { err = errors.Join(err, archive.Close()) }()
	names := []string{"manifest.json"}
	create := func(name string) (io.Writer, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entry, err := archive.Create(name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
		return entry, nil
	}
	files := make(map[string][]byte)
	if opts.Settings {
		if err := snapshot.settings(ctx, files, opts.Redact); err != nil {
			return err
		}
	}
	if opts.Logs || opts.Components {
		if err := snapshot.logs(ctx, create, opts, now.Add(-opts.Since), now); err != nil {
			return err
		}
	}
	if opts.Components {
		if err := encodeFile(files, "components.json", snapshot.components); err != nil {
			return err
		}
		if err := snapshot.environment(ctx, files); err != nil {
			return err
		}
	}
	// 2. 日志已逐条压缩，剩余配置和环境按文件名写入；清单包含自身。
	remaining := make([]string, 0, len(files))
	for name := range files {
		remaining = append(remaining, name)
	}
	slices.Sort(remaining)
	for _, name := range remaining {
		entry, err := create(name)
		if err != nil {
			return err
		}
		if _, err := entry.Write(files[name]); err != nil {
			return err
		}
	}
	slices.Sort(names)
	if err := encodeFile(files, "manifest.json", manifest{formatVersion, now, opts.Since.String(), now.Add(-opts.Since), now,
		opts.Logs, opts.Settings, opts.Components, opts.Redact, names}); err != nil {
		return err
	}
	entry, err := archive.Create("manifest.json")
	if err != nil {
		return err
	}
	_, err = entry.Write(files["manifest.json"])
	return err
}

func encodeFile(files map[string][]byte, name string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	files[name] = append(raw, '\n')
	return nil
}

func (s snapshot) settings(ctx context.Context, files map[string][]byte, redact bool) error {
	raw, err := os.ReadFile(filepath.Join(s.DataDir, "core_config.json"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		value, err := decodeJSON(raw)
		if err != nil {
			return fmt.Errorf("read Core settings: %w", err)
		}
		if redact {
			maskCredentials(value)
			root, _ := value.(map[string]any)
			adapters, _ := root["adapters"].(map[string]any)
			for adapterID, value := range adapters {
				adapter, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid adapter configuration for %s: expected object", adapterID)
				}
				owner, _ := adapter["package"].(string)
				var schema []byte
				for _, c := range s.components {
					if c.Role == "adapter" && c.install.AdapterID == adapterID && (owner == "" || owner == c.ID) {
						schema, err = s.schema(ctx, c, true)
						if err != nil {
							return err
						}
						break
					}
				}
				if err := maskAdapter(adapter, schema); err != nil {
					return err
				}
			}
		}
		if err := encodeFile(files, "settings/core_config.json", value); err != nil {
			return err
		}
	}
	for _, c := range s.components {
		if c.config == nil {
			continue
		}
		value, err := decodeJSON(c.config)
		if err != nil {
			return err
		}
		value, err = s.references(ctx, c.ID, value, redact)
		if err != nil {
			return err
		}
		if redact {
			schema, err := s.schema(ctx, c, false)
			if err != nil {
				return err
			}
			if err := maskSchema(value, schema); err != nil {
				return err
			}
			for _, path := range c.paths {
				parts := strings.Split(path, ".")
				for i, part := range parts {
					parts[i] = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "."), "~0", "~")
				}
				maskPath(value, parts)
			}
		}
		if err := encodeFile(files, "settings/plugins/"+c.ID+".json", value); err != nil {
			return err
		}
	}
	return nil
}

func (s snapshot) logs(ctx context.Context, create func(string) (io.Writer, error), opts Options, from, to time.Time) error {
	dir := filepath.Join(s.DataDir, "logs")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	lastErrors := make(map[string]time.Time)
	for _, entry := range entries {
		match := logName.FindStringSubmatch(entry.Name())
		if match == nil || !entry.Type().IsRegular() {
			continue
		}
		date, err := time.ParseInLocation(dateLayout, match[1], time.Local)
		if err != nil {
			return err
		}
		if date.After(to) || (!opts.Components && date.AddDate(0, 0, 1).Before(from)) {
			continue
		}
		f, err := os.Open(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
		var output io.Writer
		appendLog := func(text string) error {
			if output == nil {
				var err error
				output, err = create("logs/" + entry.Name())
				if err != nil {
					return err
				}
			}
			_, err := io.WriteString(output, text)
			return err
		}
		readErr := s.readLog(ctx, f, date, opts, from, to, lastErrors, appendLog)
		if err := errors.Join(readErr, f.Close()); err != nil {
			return fmt.Errorf("read log %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (s snapshot) readLog(ctx context.Context, r io.Reader, date time.Time, opts Options, from, to time.Time, lastErrors map[string]time.Time, appendLog func(string) error) error {
	reader := bufio.NewReader(r)
	var record strings.Builder
	var stamp time.Time
	flush := func() error {
		text := record.String()
		if text == "" {
			return nil
		}
		_, header, _ := strings.Cut(text, "] [")
		label, body, _ := strings.Cut(header, "] ")
		if opts.Redact {
			text = redactText(text)
		}
		if opts.Components && (strings.HasPrefix(body, "[ERRO]") || strings.HasPrefix(body, "[CRITICAL]")) {
			for i := range s.components {
				c := &s.components[i]
				plugin := label == "Plug:"+strings.TrimPrefix(c.ID, "alonabot_plugin_")
				adapter := c.Role == "adapter" && (label == "Adpt:"+c.install.AdapterID || strings.HasPrefix(label, "Adpt:"+c.install.AdapterID+"/"))
				core := label == "Core" && (strings.Contains(body, " plugin_id="+c.ID+" ") || strings.HasSuffix(strings.TrimSpace(body), " plugin_id="+c.ID))
				if (plugin || adapter || core) && !stamp.Before(lastErrors[c.ID]) {
					c.LastError = strings.TrimSuffix(text, "\n")
					lastErrors[c.ID] = stamp
				}
			}
		}
		if opts.Logs && !stamp.Before(from) && !stamp.After(to) {
			if err := appendLog(text); err != nil {
				return err
			}
		}
		record.Reset()
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}
		if line != "" {
			if !strings.HasPrefix(line, "    ") {
				if err := flush(); err != nil {
					return err
				}
				clock, _, found := strings.Cut(strings.TrimPrefix(line, "["), "]")
				if !found {
					return errors.New("invalid log record header")
				}
				var err error
				stamp, err = time.ParseInLocation(dateLayout+" 15:04:05.000", date.Format(dateLayout)+" "+clock, time.Local)
				if err != nil {
					return err
				}
			} else if record.Len() == 0 {
				return errors.New("log continuation without a record")
			}
			record.WriteString(line)
		}
		if errors.Is(readErr, io.EOF) {
			return flush()
		}
	}
}

func (s snapshot) environment(ctx context.Context, files map[string][]byte) error {
	var size int64
	err := filepath.WalkDir(s.DataDir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}
	container := os.Getenv("container") != ""
	for _, name := range []string{"/.dockerenv", "/run/.containerenv"} {
		if _, err := os.Stat(name); err == nil {
			container = true
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return encodeFile(files, "environment.json", struct {
		Version      string `json:"alonabot_version"`
		OS           string `json:"os"`
		Arch         string `json:"arch"`
		GoVersion    string `json:"go_version"`
		Container    bool   `json:"container"`
		DataDirBytes int64  `json:"data_dir_bytes"`
	}{s.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), container, size})
}
