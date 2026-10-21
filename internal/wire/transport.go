package wire

import (
	"bufio"
	"errors"
	"io"
	"net"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
)

// Conn 是一条帧连接：唯一的 reader 与唯一的串行 writer。TCP 与 stdio 共用同一实现。
type Conn struct {
	r      *bufio.Reader
	w      *Writer
	closer io.Closer
}

// NewConn 由读端、写端与关闭函数组成连接。stdio 时 Core 侧为子进程的 stdout/stdin 管道，
// 插件侧为 os.Stdin/os.Stdout（stdout 只传帧，普通日志必须写 stderr）。
func NewConn(r io.Reader, w io.Writer, closer io.Closer) *Conn {
	return &Conn{r: bufio.NewReaderSize(r, 64<<10), w: NewWriter(w), closer: closer}
}

// NewStreamConn 用同一个双向流（例如 net.Conn）构造连接。
func NewStreamConn(rwc io.ReadWriteCloser) *Conn { return NewConn(rwc, rwc, rwc) }

// ReadFrame 读取下一帧；只能由该连接唯一的读循环调用。
func (c *Conn) ReadFrame() (*pb.Envelope, error) { return ReadFrame(c.r) }

// WriteFrame 串行写出一帧，可被多个 goroutine 并发调用。
func (c *Conn) WriteFrame(env *pb.Envelope) error { return c.w.WriteFrame(env) }

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
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	if a, ok := ln.Addr().(*net.TCPAddr); !ok || !a.IP.IsLoopback() || a.Port == 0 {
		_ = ln.Close()
		return nil, errors.New("Listener 未绑定到 loopback 的非零端口")
	}
	return ln, nil
}
