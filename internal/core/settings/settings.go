// Package settings 读取 Core 启动设置，并在首次启动时创建缺省 alonabot.yaml。
//
// 调用 [Load] 前由装配方取得 data_dir 的 core.lock；命令行覆盖由装配方应用，文件内容保持不变。
// [Settings] 与 [Defaults] 遵循 CONTRACTS §15.1，日志文件上限遵循 §14.6。
package settings

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"go.yaml.in/yaml/v3"
)

// FileName 指定 data_dir 中的 Core 设置文件名。
const FileName = "alonabot.yaml"

const defaultPort = "7650"
const defaultListen = "127.0.0.1:" + defaultPort

// DNS文本长度扣除可选的终止点，遵循net解析器采用的RFC 1035线格式上限。
const (
	maxDNSNameBytes  = 253
	maxDNSLabelBytes = 63
)

// Settings 保存本次运行的启动设置；大小以字节计，TrustedProxies 保存可信反代的 CIDR 范围。
type Settings struct {
	Listen         string
	AdminBaseURL   string
	PluginBaseURL  string
	TrustedProxies []netip.Prefix
	LogLevel       slog.Level
	LogFormat      string
	LogFile        struct {
		Enabled           bool
		MaxSize, MaxTotal int64 // 字节
		MaxAgeDays        int
	}
	Proxy string
}

// Defaults 返回 CONTRACTS §15.1 的启动缺省值。
func Defaults() Settings {
	s := Settings{Listen: defaultListen, AdminBaseURL: "http://" + defaultListen, PluginBaseURL: "http://localhost:" + defaultPort, LogLevel: slog.LevelInfo, LogFormat: string(logging.FormatText)}
	s.LogFile.Enabled = true
	s.LogFile.MaxSize = logging.DefaultMaxFileSize
	s.LogFile.MaxTotal = logging.DefaultMaxTotalSize
	s.LogFile.MaxAgeDays = logging.DefaultMaxAgeDays
	return s
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

// Load 读取启动设置；文件不存在时以 0600 权限创建缺省内容。
// 未知键、类型与值错误包含完整键名和 YAML 行号。
func Load(dataDir string) (Settings, error) {
	name := filepath.Join(dataDir, FileName)
	content, err := os.ReadFile(name)
	if errors.Is(err, os.ErrNotExist) {
		s := Defaults()
		content, err = yaml.Marshal(map[string]any{
			"listen": s.Listen, "admin_base_url": s.AdminBaseURL, "plugin_base_url": s.PluginBaseURL,
			"trusted_proxies": []string{}, "proxy": s.Proxy,
			"log": map[string]any{"level": strings.ToLower(s.LogLevel.String()), "format": s.LogFormat,
				"file": map[string]any{"enabled": s.LogFile.Enabled, "max_size": fmt.Sprintf("%dMiB", s.LogFile.MaxSize>>20), "max_total": fmt.Sprintf("%dGiB", s.LogFile.MaxTotal>>30), "max_age_days": s.LogFile.MaxAgeDays}},
		})
		if err != nil {
			return Settings{}, fmt.Errorf("encode default settings: %w", err)
		}
		if err := os.WriteFile(name, content, 0o600); err != nil {
			return Settings{}, fmt.Errorf("create %s: %w", name, err)
		}
		return s, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read %s: %w", name, err)
	}
	// 1. 保留 YAML 节点的位置，类型与语义错误都指向用户填写的键。
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&document); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", name, err)
	}
	if len(document.Content) == 0 {
		return Settings{}, fmt.Errorf("%s: settings: line 1: expected a mapping", name)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return Settings{}, fmt.Errorf("%s: %w", name, err)
		}
		return Settings{}, fmt.Errorf("%s: settings: line %d: expected a single YAML document", name, extra.Content[0].Line)
	}
	s := Defaults()
	locations := map[string]*yaml.Node{"": document.Content[0]}
	if err := readMapping(document.Content[0], "", &s, locations); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", name, err)
	}
	// 2. 联合校验 Base URL、监听安全边界与日志文件限额。
	admin, err := baseURL(s.AdminBaseURL)
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "admin_base_url", err))
	}
	plugin, err := baseURL(s.PluginBaseURL)
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "plugin_base_url", err))
	}
	if admin.Scheme != plugin.Scheme || effectivePort(admin) != effectivePort(plugin) {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "plugin_base_url", errors.New("must use the same scheme and effective port as admin_base_url")))
	}
	if strings.EqualFold(admin.Hostname(), plugin.Hostname()) {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "plugin_base_url", errors.New("must use a different hostname from admin_base_url")))
	}
	host, _, _ := net.SplitHostPort(s.Listen) // listen 的语法已在 readValue 中校验。
	ip, _ := netip.ParseAddr(host)
	if !ip.IsLoopback() && !strings.EqualFold(strings.TrimSuffix(host, "."), "localhost") && admin.Scheme != "https" && len(s.TrustedProxies) == 0 {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "listen", errors.New("non-loopback listen requires HTTPS admin_base_url or trusted_proxies")))
	}
	if s.LogFile.MaxSize > s.LogFile.MaxTotal {
		return Settings{}, fmt.Errorf("%s: %w", name, at(locations, "log.file.max_size", errors.New("exceeds log.file.max_total")))
	}
	return s, nil
}

func readMapping(n *yaml.Node, prefix string, s *Settings, locations map[string]*yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return at(locations, prefix, errors.New("expected a mapping"))
	}
	seen := make(map[string]bool)
	for i := 0; i < len(n.Content); i += 2 {
		key, value := n.Content[i], n.Content[i+1]
		name := key.Value
		if prefix != "" {
			name = prefix + "." + name
		}
		locations[name] = key
		if key.Tag != "!!str" {
			return at(locations, name, errors.New("expected a string key"))
		}
		if strings.Contains(key.Value, ".") {
			return at(locations, name, errors.New("unknown key; use nested YAML mappings"))
		}
		if seen[name] {
			return at(locations, name, errors.New("duplicate key"))
		}
		seen[name] = true
		if name == "log" || name == "log.file" {
			if err := readMapping(value, name, s, locations); err != nil {
				return err
			}
			continue
		}
		if err := readValue(value, name, s); err != nil {
			return at(locations, name, err)
		}
	}
	return nil
}

func readValue(n *yaml.Node, key string, s *Settings) error {
	switch key {
	case "listen", "admin_base_url", "plugin_base_url", "log.level", "log.format", "log.file.max_size", "log.file.max_total", "proxy":
		if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
			return errors.New("expected a string")
		}
	case "log.file.enabled":
		if n.Tag != "!!bool" {
			return errors.New("expected a boolean")
		}
		return n.Decode(&s.LogFile.Enabled)
	case "log.file.max_age_days":
		if n.Tag != "!!int" {
			return errors.New("expected an integer")
		}
		if err := n.Decode(&s.LogFile.MaxAgeDays); err != nil {
			return err
		}
		if s.LogFile.MaxAgeDays <= 0 {
			return errors.New("must be positive")
		}
		return nil
	case "trusted_proxies":
		if n.Kind != yaml.SequenceNode {
			return errors.New("expected a sequence of CIDR strings")
		}
		s.TrustedProxies = nil
		for _, item := range n.Content {
			if item.Tag != "!!str" {
				return fmt.Errorf("line %d: expected a CIDR string", item.Line)
			}
			prefix, err := netip.ParsePrefix(item.Value)
			if err != nil {
				return fmt.Errorf("line %d: invalid CIDR: %w", item.Line, err)
			}
			s.TrustedProxies = append(s.TrustedProxies, prefix)
		}
		return nil
	default:
		return errors.New("unknown key")
	}
	switch key {
	case "listen":
		if err := ValidateListen(n.Value); err != nil {
			return err
		}
		s.Listen = n.Value
	case "admin_base_url":
		s.AdminBaseURL = n.Value
	case "plugin_base_url":
		s.PluginBaseURL = n.Value
	case "log.level":
		switch n.Value {
		case "debug":
			s.LogLevel = slog.LevelDebug
		case "info":
			s.LogLevel = slog.LevelInfo
		case "warn":
			s.LogLevel = slog.LevelWarn
		case "error":
			s.LogLevel = slog.LevelError
		default:
			return errors.New("expected debug, info, warn or error")
		}
	case "log.format":
		if n.Value != string(logging.FormatText) && n.Value != string(logging.FormatJSON) {
			return errors.New("expected text or json")
		}
		s.LogFormat = n.Value
	case "log.file.max_size", "log.file.max_total":
		size, err := tune.ParseSize(n.Value)
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
	case "proxy":
		if n.Value != "" {
			u, err := url.Parse(n.Value)
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
		s.Proxy = n.Value
	}
	return nil
}

func at(locations map[string]*yaml.Node, key string, err error) error {
	node, ok := locations[key]
	if !ok {
		node = locations[""]
	}
	if key == "" {
		key = "settings"
	}
	return fmt.Errorf("%s: line %d: %w", key, node.Line, err)
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

func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		n, _ := strconv.Atoi(port)
		return strconv.Itoa(n)
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
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
