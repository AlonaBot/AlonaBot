// Package wire 实现 Core 与 SDK 共用的协议运行时：帧编解码、传输与 Session 调用引擎。
//
// 每帧由 4 字节 big-endian 载荷长度（不含前缀）和 Protobuf Envelope 组成，长度范围为
// 1～4,194,304 字节；TCP 与 stdio 使用相同的编码。帧格式见 spec/WIRE-FORMAT.md §1。
//
// Session 在一条已认证的连接上同时收发双向调用：按调用编号配对请求与响应，
// 执行每方向与全局的准入预算，并让响应、取消和控制帧走优先通道，不被普通请求阻塞。
// 调用与取消的规则见 spec/WIRE-FORMAT.md §4。
package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

// MaxFrame 限制单帧载荷（Envelope 编码后）的字节数。
const MaxFrame = 4 << 20

var (
	// ErrEmptyFrame 表示长度前缀为 0。
	ErrEmptyFrame = errors.New("帧长度为 0")
	// ErrFrameTooLarge 表示帧长度超过 MaxFrame。
	ErrFrameTooLarge = errors.New("帧长度超过 4 MiB")
	// ErrTruncated 表示流在一帧中途结束。
	ErrTruncated = errors.New("帧被截断")
	// ErrDecode 表示载荷不是可解码的 Envelope。
	ErrDecode = errors.New("Envelope 无法解码")
)

// 以上错误都使当前连接失败，不再猜测后续帧边界（WIRE §1）。

// ReadRaw 读取一帧的原始载荷。先校验长度再分配缓冲；流恰好在帧边界结束时返回 io.EOF。
func ReadRaw(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("%w：读取长度前缀: %v", ErrTruncated, err)
	}
	n := binary.BigEndian.Uint32(hdr[:])
	switch {
	case n == 0:
		return nil, ErrEmptyFrame
	case n > MaxFrame:
		return nil, fmt.Errorf("%w：%d 字节", ErrFrameTooLarge, n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("%w：期望 %d 字节: %v", ErrTruncated, n, err)
	}
	return buf, nil
}

// ReadFrame 读取并解码一帧 Envelope；结构合法性由 contracts/validate 另行检查。
func ReadFrame(r io.Reader) (*pb.Envelope, error) {
	buf, err := ReadRaw(r)
	if err != nil {
		return nil, err
	}
	env := &pb.Envelope{}
	if err := proto.Unmarshal(buf, env); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecode, err)
	}
	return env, nil
}

// EncodeFrame 把 Envelope 编码为带长度前缀的完整帧；超过 MaxFrame 时不产生任何字节。
func EncodeFrame(env *pb.Envelope) ([]byte, error) {
	body, err := proto.Marshal(env)
	if err != nil {
		return nil, err
	}
	switch {
	case len(body) == 0:
		return nil, ErrEmptyFrame
	case len(body) > MaxFrame:
		return nil, fmt.Errorf("%w：%d 字节", ErrFrameTooLarge, len(body))
	}
	frame := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(frame, uint32(len(body)))
	copy(frame[4:], body)
	return frame, nil
}

// Writer 串行写帧：每帧完整编码后在锁内一次写出，多个 goroutine 的帧不会交错。
type Writer struct {
	// mu 在整帧写出期间独占写端。
	mu sync.Mutex
	w  io.Writer // 底层连接的写端
}

// NewWriter 包装底层连接的写端；每个连接只应有一个 Writer。
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

// WriteFrame 编码并写出一帧。写出失败后连接状态未知，调用方应关闭连接。
func (w *Writer) WriteFrame(env *pb.Envelope) error {
	// 1. 在取得写锁前完成编码与长度校验。
	frame, err := EncodeFrame(env)
	if err != nil {
		return err
	}
	// 2. 整帧写出期间独占写端，避免并发帧交错。
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err = w.w.Write(frame)
	return err
}
