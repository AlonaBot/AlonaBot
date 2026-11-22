package validate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// 插件 ID 规则。
const (
	// MaxPluginIDLen 限定插件 ID 的最大字节数。
	MaxPluginIDLen = 64
	// MarketPrefix 标识市场插件 ID 的固定前缀。
	MarketPrefix = "alonabot-plugin-"
)

var (
	marketID = regexp.MustCompile(`^alonabot-plugin-[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	localID  = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`) // 长度由 PluginID 先行检查，避免展开有界重复。
)

// reservedNames 收录 Windows 保留设备名，它们作为路径组件在 Windows 上不可用。
var reservedNames = func() map[string]bool {
	m := map[string]bool{"con": true, "prn": true, "aux": true, "nul": true}
	for i := 1; i <= 9; i++ {
		m[fmt.Sprintf("com%d", i)] = true
		m[fmt.Sprintf("lpt%d", i)] = true
	}
	return m
}()

// PluginID 校验插件 ID。市场 ID 直接通过；不带市场前缀的安全本地 ID 通过但返回警告。
// 大写字母被拒绝：ID 以规范小写保存，不能靠大小写区分两个安装。
func PluginID(id string) (warnings []string, err error) {
	switch {
	case id == "":
		return nil, errors.New("missing plugin ID")
	case len(id) > MaxPluginIDLen:
		return nil, fmt.Errorf("plugin ID length %d exceeds %d", len(id), MaxPluginIDLen)
	case marketID.MatchString(id):
		return nil, nil
	case strings.HasPrefix(id, MarketPrefix):
		return nil, fmt.Errorf("plugin ID %q has the market prefix but does not match the market ID rule", id)
	case strings.ToLower(id) != id:
		return nil, fmt.Errorf("plugin ID %q contains uppercase letters; use lowercase", id)
	case !localID.MatchString(id):
		return nil, fmt.Errorf("plugin ID %q is not a safe slug (must start with a lowercase letter, followed by lowercase letters, digits, _ or -)", id)
	case reservedNames[id]:
		return nil, fmt.Errorf("plugin ID %q is a Windows reserved device name", id)
	}
	return []string{fmt.Sprintf("plugin ID %q lacks the market prefix %s and can only be installed as a local plugin", id, MarketPrefix)}, nil
}
