//go:build linux

// 本文件测试 Core 作为 subreaper（容器中作为 PID 1 时同理）时，KillTree 会回收过继来的孙进程僵尸，不会无限等待。
package process

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// 测试进程设为 subreaper 后，被杀的孙进程过继给它并成为僵尸；KillTree 必须回收它们并在有限时间内返回。
func TestKillTreeReapsAdoptedZombies(t *testing.T) {
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		t.Fatalf("设置 subreaper 失败: %v", err)
	}
	defer func() { _ = unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 0, 0, 0, 0) }()

	lines := make(chan string, 16)
	p, err := Start(spec("spawn", nil), func(_, stream string, line []byte) {
		if stream == "stdout" {
			lines <- string(line)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	var gpid int
	select {
	case l := <-lines:
		if _, err := fmt.Sscanf(l, "grandchild %d", &gpid); err != nil {
			t.Fatalf("辅助进程输出 %q", l)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("没有收到孙进程 pid")
	}

	done := make(chan error, 1)
	go func() { done <- p.KillTree() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		_ = unix.Kill(gpid, unix.SIGKILL)
		t.Fatal("KillTree 未返回：过继来的僵尸没有被回收")
	}
	<-p.Exited()
	if err := unix.Kill(gpid, 0); err == nil {
		t.Fatalf("孙进程 %d 仍存在（包括僵尸）", gpid)
	}
}
