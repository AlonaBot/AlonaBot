package config

// Windows 的目录持久性沿用 settings 包约定，文件内容在发布前已 Sync。
func syncDirectory(string) error { return nil }
