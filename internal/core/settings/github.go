package settings

import (
	"errors"
	"net/url"
	"strings"
)

// ValidateGitHubEndpoint 校验可选的 API 基址或下载前缀；空串表示使用官方地址。
// 下载前缀可带查询参数，调用方将原资产 URL 原样接在其后。
// API 基址决定 Release 摘要的信任根，仅应使用可信的 GitHub API 反向代理。
func ValidateGitHubEndpoint(value string, prefix bool) error {
	if value == "" {
		return nil
	}
	u, err := url.Parse(value)
	if err != nil {
		return errors.New("GitHub endpoint must be a valid HTTPS URL")
	}
	if u.Scheme != "https" || !validHost(u.Hostname()) || u.User != nil || u.Fragment != "" || (!prefix && (u.RawQuery != "" || u.ForceQuery)) {
		return errors.New("GitHub endpoint requires HTTPS, a valid host, no credentials or fragment, and no query in the API base")
	}
	if strings.HasSuffix(u.Host, ":") {
		return errors.New("GitHub endpoint port must not be empty")
	}
	if u.Port() != "" {
		return validPort(u.Port(), false)
	}
	return nil
}
