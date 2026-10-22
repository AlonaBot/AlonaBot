//go:build unix

// 本文件测试 Unix 进程组的终止：组内孙进程随 KillTree 一并消失。
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
			t.Fatalf("辅助进程输出 %q", l)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("没有收到孙进程 pid")
	}
	if err := syscall.Kill(gpid, 0); err != nil {
		t.Fatalf("孙进程应在运行: %v", err)
	}
	if err := p.KillTree(); err != nil {
		t.Fatal(err)
	}
	<-p.Exited()
	// 孙进程被 SIGKILL 后由 init/launchd 回收；轮询观察它确实消失（仅用于观察，不作为放行条件）。
	deadline := time.Now().Add(5 * time.Second)
	for syscall.Kill(gpid, 0) == nil {
		if time.Now().After(deadline) {
			t.Fatalf("孙进程 %d 在进程组被杀后仍存活", gpid)
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.WaitLogs() // 孙进程退出后日志管道的最后一个写端关闭
	if err := p.KillTree(); err != nil {
		t.Fatalf("对已空的进程组再次 KillTree 应无副作用: %v", err)
	}
}
