package pluginkit

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func checkHardlink(path string, info os.FileInfo) (err error) {
	if !info.Mode().IsRegular() {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	var details windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &details); err != nil {
		return err
	}
	if details.NumberOfLinks > 1 {
		return fmt.Errorf("resource %q is a hard link", path)
	}
	return nil
}
