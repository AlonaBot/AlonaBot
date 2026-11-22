// Package settings 读取 Core 的 core_config.json，并在首次启动时创建缺省文件。
//
// 调用 [Load] 前由装配方取得 data_dir 的 core.lock；命令行覆盖由装配方应用。
// [File] 持有未展开的 [Document] 和有效 [Settings]，副本中的环境变量已展开。
// Core 设置与 Adapter 实例按声明校正运行副本，原文件只经显式 Update 修改。
// 设置与日志文件的管理方式见 docs/deployment.md。
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// FileName 指定 data_dir 中的 Core 设置文件名。
const FileName = "core_config.json"

const (
	defaultPort     = "7650"
	defaultListen   = "127.0.0.1:" + defaultPort
	containerListen = "0.0.0.0:" + defaultPort
	defaultLocale   = "en-US"
	// 源码构建下载 Go 发行版与模块的缺省来源。
	defaultGoDownloadBase = "https://go.dev/dl/"
	defaultGoProxy        = "https://proxy.golang.org,direct"
)

// DNS 文本长度扣除可选的终止点，遵循 net 解析器采用的 RFC 1035 线格式上限。
const (
	maxDNSNameBytes  = 253
	maxDNSLabelBytes = 63
)

// Admin 按 Adapter 与外部用户 ID 绑定 Core Admin。
type Admin struct{ AdapterID, UserID string }

// String 返回可写入 admins 的用户字符串，用户 ID 中的冒号保持原样。
func (a Admin) String() string { return a.AdapterID + ":" + a.UserID }

// AdapterSettings 保存实例原文副本与最近一次准备的运行校正记录。
type AdapterSettings struct {
	Package         string // 拥有实例配置的组件 ID；空串归当前已安装的组件。
	Instances       map[string]json.RawMessage
	DefaultInstance string // 为空时仅有一个实例的 Adapter 自动使用该实例。
	Corrections     map[string][]configschema.Correction
	Unknown         map[string][]string
	Configurations  map[string]json.RawMessage // 包含 required 失败实例的校正副本，供管理修订。
}

// Settings 保存有效的 Core 设置；日志文件大小以字节计。
type Settings struct {
	Listen, BaseURL string
	TrustedProxies  []netip.Prefix
	LogLevel        slog.Level
	LogFormat       string
	LogFile         struct {
		Enabled           bool
		MaxSize, MaxTotal int64 // 字节
		MaxAgeDays        int
	}
	LogSources      map[string]string
	Proxy           string
	GitHubMirror    string // 前缀式加速地址，拼在 GitHub 下载 URL 之前；空串表示直连。
	GoDownloadBase  string // Go 发行版目录，以 / 结尾。
	GoProxy         string // 源码构建使用的 GOPROXY。
	Admins          []Admin
	CommandPrefixes []string
	Reply           Reply
	Sleep           Sleep
	DefaultLocale   string
	Adapters        map[string]AdapterSettings
	Corrections     []configschema.Correction // Core 设置的运行校正，原值不写日志。
	Unknown         []string                  // Core 声明之外的 JSON 路径。
}

// File 持有设置文档及当前有效快照；方法可并发调用。
// mu 保护文档与快照的成对访问，返回给调用者的 Settings 单独复制。
type File struct {
	*fileState
	publish func(change func() error) error // 受管视图由生命周期所有者执行写入与发布。
}

type fileState struct {
	mu       sync.RWMutex
	name     string
	document *Document
	current  Settings
}

// Defaults 返回缺省设置。
// ALONABOT_CONTAINER=1 时监听所有接口，BaseURL 保持本机地址。
func Defaults() Settings {
	s := Settings{Listen: defaultListen, BaseURL: "http://" + defaultListen,
		LogLevel:  slog.LevelInfo,
		LogFormat: string(logging.FormatText), CommandPrefixes: []string{"/"}, DefaultLocale: defaultLocale,
		GoDownloadBase: defaultGoDownloadBase, GoProxy: defaultGoProxy}
	if os.Getenv("ALONABOT_CONTAINER") == "1" {
		s.Listen = containerListen
	}
	s.LogFile.Enabled = true
	s.LogFile.MaxSize = logging.DefaultMaxFileSize
	s.LogFile.MaxTotal = logging.DefaultMaxTotalSize
	s.LogFile.MaxAgeDays = logging.DefaultMaxAgeDays
	s.Sleep = Sleep{Mode: "always_on", IdleTimeout: defaultIdleTimeout, Plugins: make(map[string]string)}
	return s
}

// Settings 返回当前设置的深副本，调用方的修改不会影响 File。
func (f *File) Settings() Settings {
	f.mu.RLock()
	defer f.mu.RUnlock()
	s := f.current
	s.Corrections = slices.Clone(s.Corrections)
	for i := range s.Corrections {
		s.Corrections[i].Old = slices.Clone(s.Corrections[i].Old)
	}
	s.Unknown = slices.Clone(s.Unknown)
	s.TrustedProxies = slices.Clone(s.TrustedProxies)
	s.LogSources = maps.Clone(s.LogSources)
	s.Sleep.Plugins = maps.Clone(s.Sleep.Plugins)
	s.Admins = slices.Clone(s.Admins)
	s.CommandPrefixes = slices.Clone(s.CommandPrefixes)
	s.Adapters = maps.Clone(s.Adapters)
	for id, adapter := range s.Adapters {
		adapter.Instances = maps.Clone(adapter.Instances)
		for instance, raw := range adapter.Instances {
			adapter.Instances[instance] = slices.Clone(raw)
		}
		adapter.Corrections = maps.Clone(adapter.Corrections)
		for instance, records := range adapter.Corrections {
			records = slices.Clone(records)
			for i := range records {
				records[i].Old = slices.Clone(records[i].Old)
			}
			adapter.Corrections[instance] = records
		}
		adapter.Unknown = maps.Clone(adapter.Unknown)
		for instance, paths := range adapter.Unknown {
			adapter.Unknown[instance] = slices.Clone(paths)
		}
		adapter.Configurations = maps.Clone(adapter.Configurations)
		for instance, raw := range adapter.Configurations {
			adapter.Configurations[instance] = slices.Clone(raw)
		}
		s.Adapters[id] = adapter
	}
	return s
}

// Load 读取并校验全部设置；文件缺失时以 0600 权限创建缺省内容。
// 配置错误包含完整键路径与 JSON 行列号，环境变量只展开整个 ${NAME} 字符串。
func Load(dataDir string) (*File, error) {
	name := filepath.Join(dataDir, FileName)
	content, err := os.ReadFile(name)
	if errors.Is(err, os.ErrNotExist) {
		content, err = defaultDocument()
		if err != nil {
			return nil, fmt.Errorf("encode default settings: %w", err)
		}
		if err := os.WriteFile(name, content, 0o600); err != nil {
			return nil, fmt.Errorf("create %s: %w", name, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	document, err := parseDocument(content)
	if err != nil {
		if !json.Valid(content) {
			slog.Log(context.Background(), logging.LevelCritical, "invalid settings JSON prevents startup", "file", name, "error", err)
		}
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	current, err := document.settings()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	logCorrections(current.Corrections)
	return &File{fileState: &fileState{name: name, document: document, current: current}}, nil
}

func defaultDocument() ([]byte, error) {
	s := Defaults()
	content, err := json.MarshalIndent(map[string]any{
		"listen": s.Listen, "base_url": s.BaseURL,
		"trusted_proxies": []string{}, "proxy": s.Proxy, "admins": []string{},
		"github_mirror": s.GitHubMirror, "go_download_base": s.GoDownloadBase, "goproxy": s.GoProxy,
		"command_prefixes": s.CommandPrefixes, "reply": s.Reply,
		"default_locale": s.DefaultLocale, "adapters": map[string]any{},
		"sleep": map[string]any{"mode": s.Sleep.Mode, "idle_timeout": s.Sleep.IdleTimeout.String(), "plugins": s.Sleep.Plugins},
		"log": map[string]any{"level": strings.ToLower(s.LogLevel.String()), "format": s.LogFormat, "sources": map[string]string{},
			"file": map[string]any{"enabled": s.LogFile.Enabled, "max_size": fmt.Sprintf("%dMiB", s.LogFile.MaxSize>>20),
				"max_total": fmt.Sprintf("%dGiB", s.LogFile.MaxTotal>>30), "max_age_days": s.LogFile.MaxAgeDays}},
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(content, '\n'), nil
}

func (d *Document) settings() (Settings, error) {
	// 1. 共享校正器只修改运行副本，原文保留供 Update 和实例准备使用。
	schema, fields, defaults, err := coreSchema()
	if err != nil {
		return Settings{}, err
	}
	// Adapter 凭证按实例 schema 展开，Core 的分组声明只解释归属与实例目录。
	adapterField := fields["adapters"]
	adapterField.Type = "raw"
	fields["adapters"] = adapterField
	raw, corrections, unknown, err := correctDeclaredConfiguration(d.root.encode(), schema, fields)
	if err != nil {
		return Settings{}, err
	}
	root, err := parseNode(raw, "")
	if err != nil {
		return Settings{}, err
	}
	s := Defaults()
	for i := range corrections {
		if corrections[i].Path == "$.adapters" {
			corrections[i].Old = nil // 整个实例目录可能携带凭证。
		}
	}
	s.Corrections, s.Unknown = corrections, unknown
	if err := readCorrectedMapping(root, defaults, fields, "", &s); err != nil {
		return Settings{}, err
	}
	// 2. 联合校验 Base URL、监听安全边界与日志文件限额。
	locations := map[string]*node{"": root}
	var locate func(*node)
	locate = func(parent *node) {
		for _, field := range parent.members {
			locations[field.value.path] = field.value
			locate(field.value)
		}
	}
	locate(root)
	if _, err := baseURL(s.BaseURL); err != nil {
		resetCoreValue(root, defaults, "base_url", &s)
	}
	host, _, _ := net.SplitHostPort(s.Listen)
	ip, _ := netip.ParseAddr(host)
	base, _ := baseURL(s.BaseURL)
	baseIP, _ := netip.ParseAddr(base.Hostname())
	containerHTTP := os.Getenv("ALONABOT_CONTAINER") == "1" && base.Scheme == "http" &&
		(baseIP.IsLoopback() || strings.EqualFold(strings.TrimSuffix(base.Hostname(), "."), "localhost"))
	if !ip.IsLoopback() && !strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") && base.Scheme != "https" && len(s.TrustedProxies) == 0 && !containerHTTP {
		resetCoreValue(root, defaults, "listen", &s)
	}
	if s.LogFile.MaxSize > s.LogFile.MaxTotal {
		resetCoreValue(root, defaults, "log.file.max_size", &s)
		if s.LogFile.MaxSize > s.LogFile.MaxTotal {
			resetCoreValue(root, defaults, "log.file.max_total", &s)
		}
	}
	if err := validateNames(s, locations); err != nil {
		// 实例目录违反命名约束时，整个 adapters 字段恢复 dict 零值；原文仍保留。
		s.Corrections = append(s.Corrections, configschema.Correction{Path: "$.adapters", Reason: "reset"})
		s.Adapters = make(map[string]AdapterSettings)
	}
	return s, nil
}

func readCorrectedMapping(root, defaults *node, fields map[string]configField, path string, s *Settings) error {
	for _, field := range root.members {
		declaration, declared := fields[field.key]
		if !declared {
			continue
		}
		name := field.key
		if path != "" {
			name = path + "." + name
		}
		if name == "log" || name == "log.file" || name == "sleep" {
			if err := readCorrectedMapping(field.value, defaults, declaration.Items, name, s); err != nil {
				return err
			}
			continue
		}
		before := *s
		if err := readValue(field.value, name, s); err != nil {
			*s = before
			if name == "adapters" {
				return err // 实例结构与名称由实例准备路径处理。
			}
			resetCoreNode(field.value, defaults, name, s)
		}
	}
	return nil
}

func resetCoreValue(root, defaults *node, path string, s *Settings) {
	for _, name := range strings.Split(path, ".") {
		root = root.members[root.index(name)].value
	}
	resetCoreNode(root, defaults, path, s)
}

func resetCoreNode(old, defaults *node, path string, s *Settings) {
	record := configschema.Correction{Path: "$." + path, Reason: "reset"}
	if path != "proxy" {
		record.Old = old.encode()
	}
	s.Corrections = append(s.Corrections, record)
	value := defaults
	for _, name := range strings.Split(path, ".") {
		value = value.members[value.index(name)].value
	}
	_ = readValue(value, path, s) // 默认文档由 Defaults 构造，已通过声明与现有解析器。
	*old = *value
}

func readValue(n *node, key string, s *Settings) error {
	switch key {
	case "sleep.mode", "sleep.idle_timeout", "sleep.plugins":
		return readSleep(n, key, s)
	case "trusted_proxies":
		values, err := n.strings()
		if err != nil {
			return err
		}
		for i, value := range values {
			prefix, err := netip.ParsePrefix(value)
			if err != nil {
				return n.elements[i].fail(fmt.Errorf("invalid CIDR: %w", err))
			}
			s.TrustedProxies = append(s.TrustedProxies, prefix)
		}
		return nil
	case "command_prefixes":
		prefixes, err := readPrefixes(n)
		if err == nil {
			s.CommandPrefixes = prefixes
		}
		return err
	case "reply":
		return readReply(n, &s.Reply)
	case "log.sources":
		return n.object(func(id string, value *node) error {
			level, err := value.text()
			if err != nil {
				return err
			}
			switch level {
			case "debug", "info", "warn", "error", "critical", "off", "inherit":
			default:
				return value.fail(errors.New("expected debug, info, warn, error, critical, off or inherit"))
			}
			if s.LogSources == nil {
				s.LogSources = make(map[string]string)
			}
			s.LogSources[id] = level
			return nil
		})
	case "admins":
		return readAdmins(n, s)
	case "adapters":
		return readAdapters(n, s)
	case "log.file.enabled":
		enabled, ok := n.value.(bool)
		if !ok {
			return n.fail(errors.New("expected a boolean"))
		}
		s.LogFile.Enabled = enabled
		return nil
	case "log.file.max_age_days":
		days, err := n.integer()
		if err != nil {
			return err
		}
		if days <= 0 {
			return n.fail(errors.New("must be positive"))
		}
		s.LogFile.MaxAgeDays = days
		return nil
	case "listen", "base_url", "log.level", "log.format", "log.file.max_size", "log.file.max_total", "proxy", "default_locale",
		"github_mirror", "go_download_base", "goproxy":
		value, err := n.text()
		if err != nil {
			return err
		}
		if err := readString(value, key, s); err != nil {
			return n.fail(err)
		}
		return nil
	default:
		return nil // 未声明字段保留原文，不参与有效设置。
	}
}

func readPrefixes(n *node) ([]string, error) {
	prefixes, err := n.strings()
	if err != nil {
		return nil, err
	}
	if len(prefixes) == 0 {
		return nil, n.fail(errors.New("prefix list must not be empty; use [\"\"] for commands without a prefix"))
	}
	return prefixes, nil
}

func readAdmins(n *node, s *Settings) error {
	if n.delimiter != '[' {
		return n.fail(errors.New("expected an array"))
	}
	seen := make(map[Admin]bool)
	for _, child := range n.elements {
		value, err := child.text()
		if err != nil {
			return err
		}
		adapterID, userID, _ := strings.Cut(value, ":")
		// 平台 ID 只用本地规则，替换仅用于校验。
		if _, err := validate.PluginID(strings.Replace(adapterID, "-", "_", 1)); err != nil {
			return child.fail(fmt.Errorf("invalid adapter ID: %q", adapterID))
		}
		if userID == "" {
			return child.fail(errors.New("external user ID must not be empty"))
		}
		admin := Admin{AdapterID: adapterID, UserID: userID}
		if seen[admin] {
			return child.fail(fmt.Errorf("duplicate admin: %q", value))
		}
		seen[admin] = true
		s.Admins = append(s.Admins, admin)
	}
	return nil
}

func readAdapters(n *node, s *Settings) error {
	schema, problems := configschema.Parse([]byte(adapterGroupDeclaration))
	if len(problems) != 0 {
		return fmt.Errorf("invalid Adapter group schema: %v", problems)
	}
	s.Adapters = make(map[string]AdapterSettings)
	return n.object(func(id string, value *node) error {
		path := fmt.Sprintf("$.adapters[%q]", id)
		raw := value.encode()
		if value.delimiter != '{' {
			s.Corrections = append(s.Corrections, configschema.Correction{Path: path, Reason: "reset"})
			raw = []byte("{}")
		}
		var fields map[string]configField
		if err := json.Unmarshal(schema.Canonical(), &fields); err != nil {
			return err
		}
		instances := fields["instances"]
		instances.Type = "raw"
		fields["instances"] = instances
		corrected, records, unknown, err := correctDeclaredConfiguration(raw, schema, fields)
		if err != nil {
			return err
		}
		for i := range records {
			records[i].Path = path + strings.TrimPrefix(records[i].Path, "$")
			if records[i].Path == path+".instances" {
				records[i].Old = nil
			}
		}
		s.Corrections = append(s.Corrections, records...)
		for _, fieldPath := range unknown {
			s.Unknown = append(s.Unknown, path+strings.TrimPrefix(fieldPath, "$"))
		}
		var group struct {
			Package         string                     `json:"package"`
			DefaultInstance string                     `json:"default_instance"`
			Instances       map[string]json.RawMessage `json:"instances"`
		}
		if err := json.Unmarshal(corrected, &group); err != nil {
			return err
		}
		if group.Package != "" {
			if _, err := validate.PluginID(group.Package); err != nil {
				old, _ := json.Marshal(group.Package)
				s.Corrections = append(s.Corrections, configschema.Correction{Path: path + ".package", Reason: "reset", Old: old})
				group.Package = ""
			}
		}
		if group.DefaultInstance != "" {
			if _, found := group.Instances[group.DefaultInstance]; !found {
				old, _ := json.Marshal(group.DefaultInstance)
				s.Corrections = append(s.Corrections, configschema.Correction{Path: path + ".default_instance", Reason: "reset", Old: old})
				group.DefaultInstance = ""
			}
		}
		s.Adapters[id] = AdapterSettings{Package: group.Package, DefaultInstance: group.DefaultInstance, Instances: group.Instances}
		return nil
	})
}

func baseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || !validHost(u.Hostname()) {
		return nil, errors.New("expected an absolute http or https URL with a valid hostname or IP address")
	}
	if strings.HasSuffix(u.Host, ":") {
		return nil, errors.New("port must not be empty")
	}
	if u.Port() != "" {
		if err := validPort(u.Port(), false); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func validPort(port string, allowZero bool) error {
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	if n == 0 && !allowZero {
		return errors.New("port must be nonzero")
	}
	return nil
}

func validHost(host string) bool {
	if _, err := netip.ParseAddr(host); err == nil {
		return true
	}
	host = strings.TrimSuffix(host, ".")
	if host == "" || len(host) > maxDNSNameBytes {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > maxDNSLabelBytes || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if c != '-' && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
				return false
			}
		}
	}
	return true
}

func readString(value, key string, s *Settings) error {
	switch key {
	case "listen":
		if err := ValidateListen(value); err != nil {
			return err
		}
		s.Listen = value
	case "base_url":
		s.BaseURL = value
	case "log.level":
		switch value {
		case "debug":
			s.LogLevel = slog.LevelDebug
		case "info":
			s.LogLevel = slog.LevelInfo
		case "warn":
			s.LogLevel = slog.LevelWarn
		case "error":
			s.LogLevel = slog.LevelError
		case "critical":
			s.LogLevel = logging.LevelCritical
		default:
			return errors.New("expected debug, info, warn, error or critical")
		}
	case "log.format":
		if value != string(logging.FormatText) && value != string(logging.FormatJSON) {
			return errors.New("expected text or json")
		}
		s.LogFormat = value
	case "log.file.max_size", "log.file.max_total":
		size, err := tune.ParseSize(value)
		if err != nil {
			return err
		}
		if size <= 0 {
			return errors.New("must be positive")
		}
		if key == "log.file.max_size" {
			s.LogFile.MaxSize = size
		} else {
			s.LogFile.MaxTotal = size
		}
	case "default_locale":
		if value == "" {
			return errors.New("must not be empty")
		}
		s.DefaultLocale = value
	case "proxy":
		if value != "" {
			u, err := url.Parse(value)
			if err != nil {
				return err
			}
			if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "socks5" {
				return errors.New("expected an http, https or socks5 URL")
			}
			if !validHost(u.Hostname()) {
				return errors.New("proxy URL requires a valid hostname or IP address")
			}
			if strings.HasSuffix(u.Host, ":") {
				return errors.New("port must not be empty")
			}
			if u.Port() != "" {
				if err := validPort(u.Port(), false); err != nil {
					return err
				}
			}
		}
		s.Proxy = value
	case "github_mirror":
		if value != "" {
			if err := ValidateURLPrefix(value); err != nil {
				return err
			}
		}
		s.GitHubMirror = value
	case "go_download_base":
		if err := ValidateURLPrefix(value); err != nil {
			return err
		}
		s.GoDownloadBase = value
	case "goproxy":
		if err := ValidateGoProxy(value); err != nil {
			return err
		}
		s.GoProxy = value
	}
	return nil
}

// ValidateURLPrefix 校验 github_mirror 与 go_download_base 的前缀式地址：
// https、主机名有效、以 / 结尾，不带查询、片段或认证信息，直接与目标路径拼接。
// 源码 tarball 没有摘要，Go 发行版的摘要与归档来自同一前缀，因此不接受明文 http。
func ValidateURLPrefix(value string) error {
	u, err := baseURL(value)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return errors.New("URL prefix must use https")
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("URL prefix must not contain credentials, a query or a fragment")
	}
	if !strings.HasSuffix(value, "/") {
		return errors.New("URL prefix must end with /")
	}
	return nil
}

// ValidateGoProxy 校验 GOPROXY 列表：以逗号或竖线分隔，每项是 direct、off 或 http(s) URL。
func ValidateGoProxy(value string) error {
	for _, entry := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '|' }) {
		if entry == "direct" || entry == "off" {
			continue
		}
		if _, err := baseURL(entry); err != nil {
			return fmt.Errorf("invalid GOPROXY entry %q: %w", entry, err)
		}
	}
	if strings.Trim(value, ",|") == "" {
		return errors.New("must not be empty")
	}
	return nil
}

func at(locations map[string]*node, key string, err error) error {
	n, ok := locations[key]
	if !ok {
		n = locations[""]
	}
	copy := *n
	copy.path = key
	return copy.fail(err)
}

// ValidateListen 校验 host:port 形式的监听地址；host 可以为空，端口可以是 0。
// 返回的规则错误不含设置键名、行号或命令行参数名，由调用方添加来源。
func ValidateListen(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if err := validPort(port, true); err != nil {
		return err
	}
	if host != "" && !validHost(host) {
		return errors.New("invalid hostname or IP address")
	}
	return nil
}
