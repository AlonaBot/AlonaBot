//go:build unix

package process

import (
	"errors"
	"os/exec"
	"syscall"
)

// tree 是 Unix 进程组：子进程在 Start 时成为新进程组的组长，孙进程默认留在组内。
type tree struct{ pgid int }

func prepareTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func attachTree(cmd *exec.Cmd) (*tree, error) {
	return &tree{pgid: cmd.Process.Pid}, nil
}

func (t *tree) kill() error {
	err := syscall.Kill(-t.pgid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil // 组内已没有进程
	}
	return err
}

func (t *tree) release() error { return nil }
