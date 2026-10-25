// Package process 启动与回收插件的 Native 子进程。
//
// 子进程只继承一小组白名单环境变量，标准输出与标准错误按行转成带插件 ID 的日志。
// 进程树在 Unix 上用进程组管理，在 Windows 上用 Job Object 管理，停止时整棵树一起结束，
// 并确认已经清空后才返回。
//
// 这是对合作型子进程的管理：它不能阻止恶意进程脱离进程组，也不是安全沙箱。
// 规则见 spec/CONTRACTS-v0.1.md §7.1 与 §12.2。
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
	"time"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// inheritedEnv 列出从 Core 环境继承到插件的变量；其余变量（含管理凭证）一律不传。
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
	PluginID string            // 插件 ID，用于日志前缀与错误信息
	Path     string            // 可执行文件绝对路径
	Dir      string            // 工作目录（插件目录）；不构成文件访问限制
	Args     []string          // 命令行参数，不含 argv[0]
	Env      map[string]string // 会话变量（wire.Env*），叠加在白名单继承变量之上
	// Stdio 为 true 时 stdin/stdout 作为协议传输，只有 stderr 是日志。
	Stdio bool
}

// Process 管理一个插件进程及其进程树。
type Process struct {
	PluginID string
	Stdin    io.WriteCloser // stdio 模式下 Core 写协议帧的管道；TCP 模式为 nil
	Stdout   io.ReadCloser  // stdio 模式下 Core 读协议帧的管道；TCP 模式为 nil

	cmd  *exec.Cmd // 领头进程
	tree *tree     // 进程组或 Job Object
	// exited 在领头进程被 Wait 回收后关闭。
	exited chan struct{}
	// logs 等待各日志转发 goroutine 结束。
	logs sync.WaitGroup

	// mu 保护 state 与 err。
	mu    sync.Mutex
	state *os.ProcessState // 领头进程的退出状态；退出前为 nil
	err   error            // cmd.Wait 的返回值
}

// Start 以精简环境启动插件进程，并把它放进新的进程组/Job Object。
func Start(spec Spec, logf LogFunc) (*Process, error) {
	if logf == nil {
		logf = StderrLog
	}
	// 1. 准备命令：精简环境，并让子进程进入新的进程组（Windows 在启动后加入 Job）。
	cmd := exec.Command(spec.Path, spec.Args...)
	cmd.Dir = spec.Dir
	cmd.Env = buildEnv(spec.Env)
	prepareTree(cmd)

	p := &Process{PluginID: spec.PluginID, cmd: cmd, exited: make(chan struct{})}
	var parentEnds, childEnds []*os.File
	// closeAll 关闭一组管道端，用于失败路径的清理。
	closeAll := func(fs []*os.File) {
		for _, f := range fs {
			_ = f.Close()
		}
	}
	// pipe 创建一对 *os.File 管道端。
	pipe := func() (r, w *os.File, err error) {
		r, w, err = os.Pipe()
		return
	}

	// 2. 自行创建管道并把写端以 *os.File 交给子进程：exec 不会额外启动复制 goroutine，
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

	// 3. 启动并建立进程树管理；失败时终止已启动的进程并关闭全部管道端。
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

	// 4. 转发日志，并在后台回收领头进程、记录退出状态。
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

// buildEnv 从 Core 环境中只继承白名单变量，再叠加会话变量。
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

// treePoll 控制确认进程树清空时的轮询间隔：Unix 无法等待非子进程退出，Windows Job 也只提供进程计数查询。
const treePoll = 10 * time.Millisecond

// KillTree 立即终止整个进程树，并在树内（含领头进程）全部消失后才返回；领头进程已退出时同样
// 终止组内剩余进程。返回 nil 才表示旧进程树已确认清空，可以启动替代实例。
func (p *Process) KillTree() error {
	if err := p.tree.kill(); err != nil {
		return err
	}
	_, err := p.WaitTree(time.Time{})
	return err
}

// WaitTree 等待进程树内的全部进程退出，不发送信号；deadline 为零值时一直等待。
// 到期时仍有进程返回 false；查询失败返回错误，调用方不能把它当作已清空。
func (p *Process) WaitTree(deadline time.Time) (bool, error) {
	for {
		// Core 是 PID 1 或 subreaper 时，组内孙进程会过继给 Core，死后成为只能由 Core 回收的僵尸；
		// 僵尸仍算组内进程，不回收就永远等不到清空。只在领头进程已被 cmd.Wait 回收后才回收，
		// 避免抢走领头进程的退出状态。
		select {
		case <-p.exited:
			p.tree.reap()
		default:
		}
		empty, err := p.tree.empty()
		if err != nil || empty {
			return empty, err
		}
		if !deadline.IsZero() && !time.Now().Before(deadline) {
			return false, nil
		}
		time.Sleep(treePoll)
	}
}

// Release 释放进程树句柄（Windows 关闭 Job Object 句柄时会终止仍在其中的进程），
// 并关闭 stdio 模式下 Core 持有的协议管道端；重复调用不报错。
func (p *Process) Release() error {
	var errs []error
	for _, c := range []io.Closer{p.Stdin, p.Stdout} {
		if c != nil {
			if err := c.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(append(errs, p.tree.release())...)
}

// LockRuntime 供 Core 在旧进程树确认退出后临时持有 <pluginDir>/runtime.lock（例如复制受管 SQLite）；
// 完成后必须先释放，再启动新的 SDK 进程，不能持锁等待 SDK 启动。
func LockRuntime(pluginDir string) (*oslock.Lock, error) {
	return oslock.TryLock(filepath.Join(pluginDir, wire.RuntimeLockFile))
}
