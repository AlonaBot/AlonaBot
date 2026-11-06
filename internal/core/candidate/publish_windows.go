//go:build windows

package candidate

import "golang.org/x/sys/windows"

// publishDatabase 使用写穿移动发布文件名，避免对 Windows 目录句柄执行 Sync。
func publishDatabase(from, to string) error {
	source, err := windows.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	destination, err := windows.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(source, destination, windows.MOVEFILE_WRITE_THROUGH)
}
