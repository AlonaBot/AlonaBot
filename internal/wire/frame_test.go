package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/iotest"

	pb "github.com/AlonaBot/alonabot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

func header(n uint32) []byte {
	var h [4]byte
	binary.BigEndian.PutUint32(h[:], n)
	return h[:]
}

func TestReadRawBoundaries(t *testing.T) {
	max := bytes.Repeat([]byte{0xAB}, MaxFrame)
	cases := []struct {
		name    string
		stream  []byte
		wantErr error
		wantLen int
	}{
		{"1 字节", append(header(1), 0x08), nil, 1},
		{"恰好 4 MiB", append(header(MaxFrame), max...), nil, MaxFrame},
		{"长度 0", header(0), ErrEmptyFrame, 0},
		// 只给出前缀：超限必须在读取载荷之前就被拒绝，不能先分配 4 MiB+1 的缓冲再报截断。
		{"4 MiB+1", header(MaxFrame + 1), ErrFrameTooLarge, 0},
		{"载荷截断", append(header(100), make([]byte, 50)...), ErrTruncated, 0},
		{"前缀截断", []byte{0, 0}, ErrTruncated, 0},
		{"空流", nil, io.EOF, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ReadRaw(bytes.NewReader(c.stream))
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("err = %v，期望 %v", err, c.wantErr)
			}
			if len(got) != c.wantLen {
				t.Fatalf("长度 = %d，期望 %d", len(got), c.wantLen)
			}
		})
	}
}

// 分多次 Read 才凑齐一帧（每次只给 1 字节）也必须完整读出。
func TestReadFrameOneByteReads(t *testing.T) {
	env := &pb.Envelope{ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_HEARTBEAT, SessionId: bytes.Repeat([]byte{1}, 16)}
	frame, err := EncodeFrame(env)
	if err != nil {
		t.Fatal(err)
	}
	r := iotest.OneByteReader(bytes.NewReader(append(frame, frame...)))
	for i := 0; i < 2; i++ {
		got, err := ReadFrame(r)
		if err != nil || !proto.Equal(got, env) {
			t.Fatalf("第 %d 帧: %v %v", i, got, err)
		}
	}
	if _, err := ReadFrame(r); !errors.Is(err, io.EOF) {
		t.Fatalf("帧边界处结束应为 io.EOF，实际 %v", err)
	}
}

func TestReadFrameDecodeError(t *testing.T) {
	if _, err := ReadFrame(bytes.NewReader(append(header(3), 0xFF, 0xFF, 0xFF))); !errors.Is(err, ErrDecode) {
		t.Fatalf("无法解码的载荷应返回 ErrDecode，实际 %v", err)
	}
}

// stdout 混入普通日志：日志字节被当作长度前缀，连接必须失败而不是继续猜测边界。
func TestStdoutPollutionFailsSession(t *testing.T) {
	frame, _ := EncodeFrame(&pb.Envelope{ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_HEARTBEAT})
	polluted := append([]byte("plugin started\n"), frame...)
	if _, err := ReadFrame(bytes.NewReader(polluted)); err == nil {
		t.Fatal("混入日志的 stdout 应使读取失败")
	}
}

func TestEncodeFrameTooLarge(t *testing.T) {
	env := &pb.Envelope{ProtocolMajor: 1, Payload: make([]byte, MaxFrame)}
	if _, err := EncodeFrame(env); !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("超限 Envelope 应在编码时被拒绝，实际 %v", err)
	}
	var buf bytes.Buffer
	if err := NewWriter(&buf).WriteFrame(env); !errors.Is(err, ErrFrameTooLarge) || buf.Len() != 0 {
		t.Fatalf("超限帧不得写出任何字节：err=%v len=%d", err, buf.Len())
	}
}

// 读取 contracts 的 golden 帧作为两种传输共用的 fixture。
func goldenFrames(t *testing.T) map[string]*pb.Envelope {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "..", "contracts", "testdata", "frames", "*.bin"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("找不到 golden 帧: %v", err)
	}
	out := map[string]*pb.Envelope{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		env := &pb.Envelope{}
		if err := proto.Unmarshal(b, env); err != nil {
			t.Fatal(err)
		}
		out[filepath.Base(p)] = env
	}
	return out
}

// roundTrip 经 conns 的写端发出全部 fixture，再从读端读回。
func roundTrip(t *testing.T, writer, reader *Conn, frames map[string]*pb.Envelope) map[string]*pb.Envelope {
	t.Helper()
	names := make([]string, 0, len(frames))
	for n := range frames {
		names = append(names, n)
	}
	errc := make(chan error, 1)
	go func() {
		for _, n := range names {
			if err := writer.WriteFrame(frames[n]); err != nil {
				errc <- err
				return
			}
		}
		errc <- nil
	}()
	got := map[string]*pb.Envelope{}
	for _, n := range names {
		env, err := reader.ReadFrame()
		if err != nil {
			t.Fatalf("读取 %s: %v", n, err)
		}
		got[n] = env
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	return got
}

func TestSameFixtureOverTCPAndStdio(t *testing.T) {
	frames := goldenFrames(t)

	// TCP：Core 持有 loopback Listener，插件拨号。
	ln, err := ListenLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if a := ln.Addr().(*net.TCPAddr); !a.IP.Equal(net.IPv4(127, 0, 0, 1)) || a.Port == 0 {
		t.Fatalf("Listener 地址 %v", a)
	}
	accepted := make(chan net.Conn, 1)
	go func() {
		c, _ := ln.Accept()
		accepted <- c
	}()
	client, err := net.Dial("tcp4", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	server := <-accepted
	viaTCP := roundTrip(t, NewStreamConn(client), NewStreamConn(server), frames)
	_ = client.Close()
	_ = server.Close()

	// stdio：用 os.Pipe 模拟子进程的 stdout（插件写、Core 读）。
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	viaStdio := roundTrip(t, NewConn(nil, pw, pw), NewConn(pr, nil, pr), frames)
	_ = pw.Close()
	_ = pr.Close()

	for name, want := range frames {
		if !proto.Equal(viaTCP[name], want) || !proto.Equal(viaStdio[name], want) {
			t.Fatalf("%s 在两种传输上的往返结果不一致", name)
		}
		a, _ := EncodeFrame(viaTCP[name])
		b, _ := EncodeFrame(viaStdio[name])
		if !bytes.Equal(a, b) {
			t.Fatalf("%s 的重新编码字节在两种传输间不同", name)
		}
	}
}

// 多个 goroutine 并发写同一连接，读端每一帧都能完整解码（帧不交错）。
func TestConcurrentWritesDoNotInterleave(t *testing.T) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	w := NewConn(nil, pw, pw)
	const writers, perWriter = 16, 50
	var wg sync.WaitGroup
	for g := 0; g < writers; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				env := &pb.Envelope{ProtocolMajor: 1, Kind: pb.FrameKind_FRAME_KIND_REQUEST,
					CallId: uint64(g*perWriter + i + 1), Payload: bytes.Repeat([]byte{byte(g)}, 1000+g*37)}
				if err := w.WriteFrame(env); err != nil {
					t.Error(err)
					return
				}
			}
		}(g)
	}
	go func() { wg.Wait(); _ = pw.Close() }()

	r := NewConn(pr, nil, pr)
	seen := map[uint64]bool{}
	for {
		env, err := r.ReadFrame()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("读到损坏的帧（帧交错）: %v", err)
		}
		g := int(env.GetCallId()-1) / perWriter
		if len(env.GetPayload()) != 1000+g*37 || env.GetPayload()[0] != byte(g) {
			t.Fatalf("call %d 的载荷不完整", env.GetCallId())
		}
		seen[env.GetCallId()] = true
	}
	if len(seen) != writers*perWriter {
		t.Fatalf("收到 %d 帧，期望 %d", len(seen), writers*perWriter)
	}
}
