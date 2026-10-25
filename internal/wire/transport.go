package wire

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// readBufferSize 是帧连接读端的缓冲大小。
const readBufferSize = 64 << 10

// loopbackListenAddr 让 Core 在 IPv4 loopback 上监听随机端口。
const loopbackListenAddr = "127.0.0.1:0"

// Conn 在一条连接上读写帧：只有一个 reader 和一个串行 writer，TCP 与 stdio 共用同一实现。
type Conn struct {
	r *bufio.Reader // 带缓冲的读端，只由唯一的读循环使用
	w *Writer       // 串行写帧器，可被并发调用
	// closer 关闭底层连接；为 nil 时 Close 不做任何事。
	closer io.Closer
}

// NewConn 由读端、写端与关闭函数组成连接。stdio 时 Core 侧为子进程的 stdout/stdin 管道，
// 插件侧为 os.Stdin/os.Stdout（stdout 只传帧，普通日志必须写 stderr）。
func NewConn(r io.Reader, w io.Writer, closer io.Closer) *Conn {
	return &Conn{r: bufio.NewReaderSize(r, readBufferSize), w: NewWriter(w), closer: closer}
}

// NewStreamConn 用同一个双向流（例如 net.Conn）构造连接。
func NewStreamConn(rwc io.ReadWriteCloser) *Conn { return NewConn(rwc, rwc, rwc) }

// ReadFrame 读取下一帧；只能由该连接唯一的读循环调用。
func (c *Conn) ReadFrame() (*pb.Envelope, error) { return ReadFrame(c.r) }

// WriteFrame 串行写出一帧，可被多个 goroutine 并发调用。
func (c *Conn) WriteFrame(env *pb.Envelope) error { return c.w.WriteFrame(env) }

// WriteFrameTimeout 在底层连接支持写截止时间时（net.Conn、Unix 管道）限时写出一帧，用于失败路径上
// 尽力通知对端（例如 SESSION_ERROR），避免对端不读时永久阻塞；不支持截止时间时退化为普通写。
func (c *Conn) WriteFrameTimeout(env *pb.Envelope, d time.Duration) error {
	if dl, ok := c.w.w.(interface {
		// SetWriteDeadline 设置写截止时间，零值表示取消。
		SetWriteDeadline(time.Time) error
	}); ok {
		if err := dl.SetWriteDeadline(time.Now().Add(d)); err == nil {
			defer dl.SetWriteDeadline(time.Time{})
		}
	}
	return c.w.WriteFrame(env)
}

// Close 关闭底层连接。
func (c *Conn) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}

// ListenLoopback 直接在 IPv4 loopback 的临时端口上监听（CONTRACTS §1.1）。调用方持有 Listener
// 直到会话结束，不采用先探测空闲端口、关闭后再绑定的做法。
func ListenLoopback() (net.Listener, error) {
	ln, err := net.Listen("tcp4", loopbackListenAddr)
	if err != nil {
		return nil, err
	}
	if a, ok := ln.Addr().(*net.TCPAddr); !ok || !a.IP.IsLoopback() || a.Port == 0 {
		_ = ln.Close()
		return nil, errors.New("listener is not bound to a non-zero loopback port")
	}
	return ln, nil
}
