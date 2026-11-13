package sdk

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// levelOff 高于任何记录的级别：Core 下发 OFF 时本进程不输出任何记录。
const levelOff = slog.Level(math.MaxInt)

// logLevel 是本进程的生效日志级别；收到 Bootstrap 之前为 INFO。
var logLevel slog.LevelVar

// setupLogging 把默认 Logger 换成以 JSON 行写 stderr、带源码位置的 Logger（CONTRACTS §14.1），
// 插件用 log/slog 或 log 包写的日志都经由它交给 Core。stdio 传输时 stdout 专用于协议帧。
//
// Core 消失后再写 stdout 或 stderr 时，Go 运行时默认以 SIGPIPE 终止进程，来不及执行 OnStop（CONTRACTS §4.2）。
// 接收该信号后这些写入只返回 EPIPE；不用 signal.Ignore，因为忽略状态会被插件启动的子进程继承。
func setupLogging() {
	signal.Notify(make(chan os.Signal, 1), syscall.SIGPIPE)
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: &logLevel})))
}

// setLogLevel 按 Core 下发的级别在进程内过滤；UNSPECIFIED 按 INFO 处理（WIRE-FORMAT §3.2）。
func setLogLevel(level *slog.LevelVar, lv pb.LogLevel) {
	switch lv {
	case pb.LogLevel_LOG_LEVEL_DEBUG:
		level.Set(slog.LevelDebug)
	case pb.LogLevel_LOG_LEVEL_WARN:
		level.Set(slog.LevelWarn)
	case pb.LogLevel_LOG_LEVEL_ERROR:
		level.Set(slog.LevelError)
	case pb.LogLevel_LOG_LEVEL_OFF:
		level.Set(levelOff)
	default:
		level.Set(slog.LevelInfo)
	}
}

func (s *componentSession) changeLogLevel(in *wire.Inbound) wire.Response {
	request := &pb.LogSetLevelRequest{}
	if err := decodeInvocation(in.Payload, request); err != nil {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error())
	}
	if _, ok := pb.LogLevel_name[int32(request.Level)]; !ok {
		return dispatchError(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "unknown log level")
	}
	setLogLevel(&s.level, request.Level)
	if !s.inProcess {
		setLogLevel(&logLevel, request.Level)
	}
	return adapterResponse(&pb.LogSetLevelResponse{}, nil)
}

// logPanic 把调用边界捕获的 panic 记为一条 ERROR，堆栈放在 stack 字段（CONTRACTS §14.1）；
// 只能在 recover 所在的 defer 中调用，堆栈才包含 panic 的位置。
func logPanic(ctx context.Context, boundary string, p any) {
	componentLogger(ctx).ErrorContext(ctx, boundary+" panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
}

func (s *componentSession) configure(l launch, ack *pb.HelloAck, component string) {
	setLogLevel(&s.level, ack.GetBootstrap().GetLogLevel())
	if !s.inProcess {
		setLogLevel(&logLevel, ack.GetBootstrap().GetLogLevel())
	}
	s.source = []slog.Attr{
		slog.String("component", component), slog.String("plugin_id", l.pluginID),
		slog.String("session_id", hex.EncodeToString(ack.GetSessionId())),
		slog.String("generation_id", ack.GetBootstrap().GetGenerationId()),
	}
}
