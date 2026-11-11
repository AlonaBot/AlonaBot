package process

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// Darwin 对正在退出或仅含僵尸的进程组也返回 EPERM；此时仍由 WaitTree 等待全部回收。
func (t *tree) signalPermissionError(signalErr error) error {
	members, err := unix.SysctlKinfoProcSlice("kern.proc.pgrp", t.pgid)
	if err != nil {
		return errors.Join(signalErr, fmt.Errorf("inspect process group %d: %w", t.pgid, err))
	}
	const (
		zombie  = 5          // Darwin sys/proc.h 的 SZOMB。
		exiting = 0x00002000 // Darwin sys/proc.h 的 P_WEXIT。
	)
	for _, member := range members {
		if member.Proc.P_stat != zombie && member.Proc.P_flag&exiting == 0 {
			return signalErr
		}
	}
	return nil
}
