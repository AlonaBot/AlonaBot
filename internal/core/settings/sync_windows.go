//go:build windows

package settings

// Windows 在 Rename 后完成提交，跳过目录句柄的 Sync。
func syncDirectory(string) error { return nil }
