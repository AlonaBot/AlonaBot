//go:build unix

package process

import (
	"fmt"
	"syscall"
	"testing"
	"time"
)

// 杀掉进程组后，留在组内的孙进程也被回收（Unix 进程组；Windows 由 Job Object 负责，只做交叉编译验证）。
func TestKillTreeReapsGrandchild(t *testing.T) {
	s := spec("spawn", nil)
	lines := make(chan string, 16)
	p, err := Start(s, func(_, stream string, line []byte) {
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
	if err := syscall.Kill(gpid, 0); err != nil {
		t.Fatalf("grandchild should be running: %v", err)
	}
	if err := p.KillTree(); err != nil {
		t.Fatal(err)
	}
	<-p.Exited()
	// 孙进程被 SIGKILL 后由 init/launchd 回收；轮询观察它确实消失（仅用于观察，不作为放行条件）。
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Kill(gpid, 0) == nil {
		if time.Now().After(deadline) {
			t.Fatalf("grandchild %d still alive after the process group was killed", gpid)
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.WaitLogs() // 孙进程退出后日志管道的最后一个写端关闭
	if err := p.KillTree(); err != nil {
		t.Fatalf("KillTree on an empty process group should have no effect: %v", err)
	}
}
