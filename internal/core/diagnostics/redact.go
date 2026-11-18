package diagnostics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
)

const redacted = "[REDACTED]"

var (
	authorization        = regexp.MustCompile(`(?i)\b(Bearer|Basic)([\t ]+)[^\s"'<>\\]+`)
	urlUser              = regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://)[^/?#\s"'<>@]+@`)
	querySecret          = regexp.MustCompile(`(?i)([?&](?:token|key|secret|password)=)[^&\s"'<>\\]*`)
	telegramToken        = regexp.MustCompile(`\bbot[0-9]+:[A-Za-z0-9_-]+`)
	ipv6                 = regexp.MustCompile(`[0-9A-Fa-f]*:[0-9A-Fa-f:.]+(?:%[A-Za-z0-9_.-]+)?`)
	ipv4                 = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	environmentReference = regexp.MustCompile(`^\$\{[A-Za-z_][A-Za-z0-9_]*\}$`)
)

func redactText(value string) string {
	value = authorization.ReplaceAllString(value, "${1}${2}"+redacted)
	value = urlUser.ReplaceAllString(value, "${1}"+redacted+"@")
	value = querySecret.ReplaceAllString(value, "${1}"+redacted)
	value = telegramToken.ReplaceAllString(value, "bot"+redacted)
	publicIP := func(text string) string {
		addr, err := netip.ParseAddr(text)
		if err != nil {
			return text
		}
		addr = addr.Unmap()
		if addr.IsGlobalUnicast() && !addr.IsPrivate() && !addr.IsLoopback() && !addr.IsLinkLocalUnicast() {
			return redacted
		}
		return text
	}
	value = ipv6.ReplaceAllStringFunc(value, func(candidate string) string {
		address := strings.TrimRight(candidate, ".") // 句末标点独立保留，地址内部的 IPv4 点号仍参与解析。
		return publicIP(address) + candidate[len(address):]
	})
	return ipv4.ReplaceAllStringFunc(value, publicIP)
}

func mask(value any) any {
	switch value := value.(type) {
	case string:
		if environmentReference.MatchString(value) {
			return value
		}
	case map[string]any:
		for key, child := range value {
			value[key] = mask(child)
		}
		return value
	case []any:
		for index, child := range value {
			value[index] = mask(child)
		}
		return value
	}
	return redacted
}

func maskPath(value any, path []string) {
	if len(path) == 0 {
		return
	}
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if key != path[0] && path[0] != "*" {
				continue
			}
			if len(path) == 1 {
				value[key] = mask(child)
			} else {
				maskPath(child, path[1:])
			}
		}
	case []any:
		for index, child := range value {
			if path[0] != "*" && path[0] != fmt.Sprint(index) {
				continue
			}
			if len(path) == 1 {
				value[index] = mask(child)
			} else {
				maskPath(child, path[1:])
			}
		}
	}
}

func maskSchema(value any, raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) > 0 {
		return fmt.Errorf("invalid diagnostic config schema: %v", problems)
	}
	for _, path := range schema.SecretPaths() {
		maskPath(value, path)
	}
	return nil
}

// maskAdapter 先迁移读取副本再遮盖；schema 缺失时保留结构并遮盖全部实例值。
func maskAdapter(adapter map[string]any, raw []byte) error {
	if len(raw) == 0 {
		if instances, found := adapter["instances"]; found {
			adapter["instances"] = mask(instances)
		}
		return nil
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) > 0 {
		return fmt.Errorf("invalid diagnostic adapter schema: %v", problems)
	}
	version := 1 // Adapter 配置省略 $version 时按首版读取，与 settings 一致。
	if value, found := adapter["$version"]; found {
		number, ok := value.(json.Number)
		if !ok {
			return fmt.Errorf("invalid adapter config version: %v", value)
		}
		var err error
		version, err = strconv.Atoi(string(number))
		if err != nil {
			return err
		}
	}
	instances, _ := adapter["instances"].(map[string]any)
	for _, value := range instances {
		config, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid adapter instance config: expected object")
		}
		if err := schema.Migrate(config, version); err != nil {
			return err
		}
		for _, path := range schema.SecretPaths() {
			maskPath(config, path)
		}
	}
	if adapter != nil {
		adapter["$version"] = schema.Version()
	}
	return nil
}

func (s snapshot) references(ctx context.Context, id string, value any, redact bool) (any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch value := value.(type) {
	case map[string]any:
		if ref, found := value[secret.RefKey]; found {
			if redact {
				return redacted, nil
			}
			text, ok := ref.(string)
			if !ok {
				return nil, fmt.Errorf("invalid secret reference in %s", id)
			}
			stored, err := s.secret(ctx, id, text)
			return string(stored.Value), err
		}
		for key, child := range value {
			next, err := s.references(ctx, id, child, redact)
			if err != nil {
				return nil, err
			}
			value[key] = next
		}
	case []any:
		for index, child := range value {
			next, err := s.references(ctx, id, child, redact)
			if err != nil {
				return nil, err
			}
			value[index] = next
		}
	}
	return value, nil
}

// maskCredentials 同时遮盖 Adapter 旧配置中的凭证名与设置 URL 内的凭证。
func maskCredentials(value any) {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			name := strings.ToLower(key)
			switch name {
			case "token", "key", "secret", "password", "authorization", "access_token", "api_key", "client_secret", "admin_token":
				value[key] = mask(child)
			default:
				if text, ok := child.(string); ok {
					value[key] = urlUser.ReplaceAllString(text, "${1}"+redacted+"@")
				}
				maskCredentials(child)
			}
		}
	case []any:
		for _, child := range value {
			maskCredentials(child)
		}
	}
}
