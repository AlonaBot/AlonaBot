//go:build unix

package process

import (
	"errors"
	"os/exec"
	"syscall"
)

// tree 对应一个 Unix 进程组：子进程在 Start 时成为新进程组的组长，孙进程默认留在组内。
type tree struct {
	pgid int // 进程组 ID，等于领头进程的 pid
}

// prepareTree 让子进程在启动时成为新进程组的组长。
func prepareTree(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// attachTree 记录已启动子进程的进程组。
func attachTree(cmd *exec.Cmd) (*tree, error) {
	return &tree{pgid: cmd.Process.Pid}, nil
}

// kill 向整个进程组发送 SIGKILL；组内已没有进程时视为成功。
func (t *tree) kill() error {
	err := syscall.Kill(-t.pgid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil // 组内已没有进程
	}
	return err
}

// empty 报告进程组内是否已没有任何进程（尚未被回收的僵尸也算组内进程）。
func (t *tree) empty() (bool, error) {
	err := syscall.Kill(-t.pgid, 0)
	if errors.Is(err, syscall.ESRCH) {
		return true, nil
	}
	return false, err
}

// reap 非阻塞地回收组内已过继给本进程的僵尸；没有可回收的子进程时立即返回。
func (t *tree) reap() {
	for {
		pid, err := syscall.Wait4(-t.pgid, nil, syscall.WNOHANG, nil)
		if pid <= 0 || err != nil {
			return
		}
	}
}

// release 无需释放资源：进程组不占用句柄。
func (t *tree) release() error { return nil }
