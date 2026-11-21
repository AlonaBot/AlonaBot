package wire

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

// 读取 contracts 的 golden 帧作为两种传输共用的 fixture。
func goldenFrames(t *testing.T) map[string]*pb.Envelope {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "..", "contracts", "testdata", "frames", "*.bin"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("golden frames not found: %v", err)
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
			t.Fatalf("read %s: %v", n, err)
		}
		got[n] = env
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	return got
}

// 同一组 golden 帧经 TCP loopback 与 stdio 管道传输后逐字节一致。
func TestSameFixtureOverTCPAndStdio(t *testing.T) {
	frames := goldenFrames(t)

	// TCP：Core 持有 loopback Listener，插件拨号。
	ln, err := ListenLoopback()
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if a := ln.Addr().(*net.TCPAddr); !a.IP.Equal(net.IPv4(127, 0, 0, 1)) || a.Port == 0 {
		t.Fatalf("listener address %v", a)
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
			t.Fatalf("round trip of %s differs between the two transports", name)
		}
		a, _ := EncodeFrame(viaTCP[name])
		b, _ := EncodeFrame(viaStdio[name])
		if !bytes.Equal(a, b) {
			t.Fatalf("re-encoded bytes of %s differ between the two transports", name)
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
			t.Fatalf("read a corrupted frame (interleaved frames): %v", err)
		}
		g := int(env.GetCallId()-1) / perWriter
		if len(env.GetPayload()) != 1000+g*37 || env.GetPayload()[0] != byte(g) {
			t.Fatalf("payload of call %d is incomplete", env.GetCallId())
		}
		seen[env.GetCallId()] = true
	}
	if len(seen) != writers*perWriter {
		t.Fatalf("got %d frames, want %d", len(seen), writers*perWriter)
	}
}
