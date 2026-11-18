package manage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func dataRenameAt(from *os.File, fromName string, to *os.File, toName string) (err error) {
	var guards []*os.File
	defer func() {
		for _, file := range guards {
			err = errors.Join(err, file.Close())
		}
	}()
	source, opened, err := dataRenamePath(from)
	guards = append(guards, opened...)
	if err != nil {
		return err
	}
	target, opened, err := dataRenamePath(to)
	guards = append(guards, opened...)
	if err != nil {
		return err
	}
	oldPath, err := windows.UTF16PtrFromString(filepath.Join(source, fromName))
	if err != nil {
		return err
	}
	newPath, err := windows.UTF16PtrFromString(filepath.Join(target, toName))
	if err != nil {
		return err
	}
	err = windows.MoveFileEx(oldPath, newPath, 0)
	// Windows 对已存在的目标目录可能返回 ACCESS_DENIED，同样按名称冲突处理。
	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		if _, statErr := windows.GetFileAttributes(newPath); statErr == nil {
			return errors.Join(os.ErrExist, err)
		}
	}
	return err
}

// dataRenamePath 将固定句柄转换为 MoveFileEx 所需的路径。逐级禁止共享写入和删除，
// 避免调用期间目录被换名或改为重解析点；末端身份必须与原句柄一致。
func dataRenamePath(directory *os.File) (name string, guards []*os.File, err error) {
	const volumeNameGUID = 0x1 // 卷 GUID 路径不依赖可被重新映射的盘符。
	var buffer []uint16
	for {
		var address *uint16
		if len(buffer) != 0 {
			address = &buffer[0]
		}
		n, err := windows.GetFinalPathNameByHandle(windows.Handle(directory.Fd()), address, uint32(len(buffer)), volumeNameGUID)
		if err != nil {
			return "", guards, err
		}
		if n < uint32(len(buffer)) {
			name = windows.UTF16ToString(buffer[:n])
			break
		}
		buffer = make([]uint16, n+1)
	}

	// 从卷根固定全部路径段，打开重解析点自身并拒绝链接。
	current := filepath.VolumeName(name) + `\`
	parts := append([]string{""}, strings.Split(strings.TrimPrefix(name, current), `\`)...)
	for _, part := range parts {
		current = filepath.Join(current, part)
		path, err := windows.UTF16PtrFromString(current)
		if err != nil {
			return "", guards, err
		}
		handle, err := windows.CreateFile(path, windows.FILE_READ_ATTRIBUTES, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING,
			windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
		if err != nil {
			return "", guards, err
		}
		file := os.NewFile(uintptr(handle), current)
		guards = append(guards, file)
		var info windows.ByHandleFileInformation
		if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
			return "", guards, err
		}
		if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
			return "", guards, invalid("symbolic link is not allowed: " + current)
		}
	}
	checked, err := directory.Stat()
	if err != nil {
		return "", guards, err
	}
	opened, err := guards[len(guards)-1].Stat()
	if err != nil {
		return "", guards, err
	}
	if !os.SameFile(checked, opened) {
		return "", guards, invalid("directory changed while preparing rename: " + name)
	}
	return name, guards, nil
}
