//go:build linux

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
		t.Fatalf("set subreaper: %v", err)
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
			t.Fatalf("helper process output %q", l)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("did not receive the grandchild pid")
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
		t.Fatal("KillTree did not return: adopted zombies were not reaped")
	}
	<-p.Exited()
	if err := unix.Kill(gpid, 0); err == nil {
		t.Fatalf("grandchild %d still exists (including as a zombie)", gpid)
	}
}
