// Package process 启动与回收插件 Native 子进程（CONTRACTS §7.1、§12.2）：精简环境、
// 带插件 ID 的日志、进程树管理（Unix 进程组 / Windows Job Object）与退出状态。
// 这是合作型子进程管理，不能阻止恶意 detach，也不是安全沙箱。
package process

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/AlonaBot/alonabot/contracts/manifest"
	"github.com/AlonaBot/alonabot/internal/oslock"
	"github.com/AlonaBot/alonabot/internal/wire"
)

// inheritedEnv 是从 Core 环境继承到插件的变量白名单；其余变量（含管理凭证）一律不传。
var inheritedEnv = []string{"PATH", "HOME", "TMPDIR", "TEMP", "TMP", "SystemRoot", "LANG", "TZ"}

// NewSecret 生成本次启动的 256-bit Session Secret（十六进制）。
func NewSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Entrypoint 根据安装目录与 manifest 中的相对入口得到可执行文件路径，并检查它是普通文件。
func Entrypoint(installDir, rel string) (string, error) {
	if err := manifest.ValidateRelPath(rel); err != nil {
		return "", fmt.Errorf("entrypoint: %w", err)
	}
	p := filepath.Join(installDir, filepath.FromSlash(rel))
	fi, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("entrypoint %s 不是普通文件", p)
	}
	if runtime.GOOS != "windows" && fi.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("entrypoint %s 不可执行", p)
	}
	return p, nil
}

// LogFunc 接收带插件 ID 的一行日志；stream 为 "stdout" 或 "stderr"。
type LogFunc func(pluginID, stream string, line []byte)

// StderrLog 把插件日志写到 Core 的 stderr，前缀为插件 ID。
func StderrLog(pluginID, stream string, line []byte) {
	fmt.Fprintf(os.Stderr, "[plugin %s %s] %s\n", pluginID, stream, line)
}

// Spec 描述一次插件进程启动。
type Spec struct {
	PluginID string
	Path     string            // 可执行文件绝对路径
	Dir      string            // 工作目录（插件目录）；不构成文件访问限制
	Args     []string          // 不含 argv[0]
	Env      map[string]string // 会话变量（wire.Env*），叠加在白名单继承变量之上
	Stdio    bool              // true：stdin/stdout 作为协议传输，只有 stderr 是日志
}

// Process 是一个受管的插件进程及其进程树。
type Process struct {
	PluginID string
	// Stdio 模式下的协议管道（Core 写 Stdin、读 Stdout）；TCP 模式为 nil。
	Stdin  io.WriteCloser
	Stdout io.ReadCloser

	cmd    *exec.Cmd
	tree   *tree
	exited chan struct{}
	logs   sync.WaitGroup

	mu    sync.Mutex
	state *os.ProcessState
	err   error
}

// Start 以精简环境启动插件进程，并把它放进新的进程组/Job Object。
func Start(spec Spec, logf LogFunc) (*Process, error) {
	if logf == nil {
		logf = StderrLog
	}
	cmd := exec.Command(spec.Path, spec.Args...)
	cmd.Dir = spec.Dir
	cmd.Env = buildEnv(spec.Env)
	prepareTree(cmd)

	p := &Process{PluginID: spec.PluginID, cmd: cmd, exited: make(chan struct{})}
	var parentEnds, childEnds []*os.File
	closeAll := func(fs []*os.File) {
		for _, f := range fs {
			_ = f.Close()
		}
	}
	pipe := func() (r, w *os.File, err error) {
		r, w, err = os.Pipe()
		return
	}

	// 自行创建管道并把写端以 *os.File 交给子进程：exec 不会额外启动复制 goroutine，
	// cmd.Wait 只等待进程本身退出，不会被仍持有管道的孙进程拖住。
	errR, errW, err := pipe()
	if err != nil {
		return nil, err
	}
	parentEnds, childEnds = append(parentEnds, errR), append(childEnds, errW)
	cmd.Stderr = errW
	var outR *os.File
	if spec.Stdio {
		inR, inW, err := pipe()
		if err != nil {
			closeAll(parentEnds)
			closeAll(childEnds)
			return nil, err
		}
		outR2, outW, err := pipe()
		if err != nil {
			closeAll(append(parentEnds, inW))
			closeAll(append(childEnds, inR))
			return nil, err
		}
		cmd.Stdin, cmd.Stdout = inR, outW
		p.Stdin, p.Stdout = inW, outR2
		parentEnds = append(parentEnds, inW, outR2)
		childEnds = append(childEnds, inR, outW)
	} else {
		r, w, err := pipe()
		if err != nil {
			closeAll(parentEnds)
			closeAll(childEnds)
			return nil, err
		}
		outR = r
		cmd.Stdout = w
		parentEnds = append(parentEnds, r)
		childEnds = append(childEnds, w)
	}

	if err := cmd.Start(); err != nil {
		closeAll(parentEnds)
		closeAll(childEnds)
		return nil, fmt.Errorf("启动插件 %s 失败: %w", spec.PluginID, err)
	}
	closeAll(childEnds) // 子进程已持有各自的副本
	t, err := attachTree(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		closeAll(parentEnds)
		return nil, fmt.Errorf("建立插件 %s 的进程树管理失败: %w", spec.PluginID, err)
	}
	p.tree = t

	p.pump("stderr", errR, logf)
	if outR != nil {
		p.pump("stdout", outR, logf)
	}
	go func() {
		err := cmd.Wait()
		p.mu.Lock()
		p.state, p.err = cmd.ProcessState, err
		p.mu.Unlock()
		close(p.exited)
	}()
	return p, nil
}

func buildEnv(extra map[string]string) []string {
	var env []string
	for _, k := range inheritedEnv {
		if v, ok := os.LookupEnv(k); ok {
			env = append(env, k+"="+v)
		}
	}
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

// pump 按行转发日志；超过 64 KiB 的行分块转发，不丢弃内容。读到 EOF（所有写端关闭）后结束。
func (p *Process) pump(stream string, r *os.File, logf LogFunc) {
	p.logs.Add(1)
	go func() {
		defer p.logs.Done()
		defer r.Close()
		br := bufio.NewReaderSize(r, 64<<10)
		for {
			line, err := br.ReadSlice('\n')
			if len(line) > 0 {
				trimmed := line
				if trimmed[len(trimmed)-1] == '\n' {
					trimmed = trimmed[:len(trimmed)-1]
				}
				logf(p.PluginID, stream, append([]byte(nil), trimmed...))
			}
			if err != nil && !errors.Is(err, bufio.ErrBufferFull) {
				return
			}
		}
	}()
}

// Pid 返回进程号。
func (p *Process) Pid() int { return p.cmd.Process.Pid }

// Exited 在进程退出（已被 Wait 回收）后关闭。
func (p *Process) Exited() <-chan struct{} { return p.exited }

// ExitState 返回退出状态；进程未退出时为 nil。
func (p *Process) ExitState() *os.ProcessState {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// WaitLogs 等待日志管道读完（所有持有写端的进程都已退出或关闭了输出）。
func (p *Process) WaitLogs() { p.logs.Wait() }

// KillTree 立即终止整个进程树（总停止截止时间到达时使用）。领头进程已退出后仍会终止组内剩余进程。
func (p *Process) KillTree() error { return p.tree.kill() }

// Release 释放进程树句柄（Windows 关闭 Job Object 句柄时会终止仍在其中的进程）。
func (p *Process) Release() error { return p.tree.release() }

// LockRuntime 供 Core 在旧进程树确认退出后临时持有 <pluginDir>/runtime.lock（例如复制受管 SQLite）；
// 完成后必须先释放，再启动新的 SDK 进程，不能持锁等待 SDK 启动。
func LockRuntime(pluginDir string) (*oslock.Lock, error) {
	return oslock.TryLock(filepath.Join(pluginDir, wire.RuntimeLockFile))
}
