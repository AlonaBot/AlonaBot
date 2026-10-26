package sdk

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime/debug"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// levelOff 高于任何记录的级别：Core 下发 OFF 时本进程不输出任何记录。
const levelOff = slog.Level(math.MaxInt)

// logLevel 是本进程的生效日志级别；收到 Bootstrap 之前为 INFO。
var logLevel slog.LevelVar

// setupLogging 把默认 Logger 换成以 JSON 行写 stderr、带源码位置的 Logger（CONTRACTS §14.1），
// 插件用 log/slog 或 log 包写的日志都经由它交给 Core。stdio 传输时 stdout 专用于协议帧。
func setupLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: &logLevel})))
}

// setLogLevel 按 Core 下发的级别在进程内过滤；UNSPECIFIED 按 INFO 处理（WIRE-FORMAT §3.2）。
func setLogLevel(lv pb.LogLevel) {
	switch lv {
	case pb.LogLevel_LOG_LEVEL_DEBUG:
		logLevel.Set(slog.LevelDebug)
	case pb.LogLevel_LOG_LEVEL_WARN:
		logLevel.Set(slog.LevelWarn)
	case pb.LogLevel_LOG_LEVEL_ERROR:
		logLevel.Set(slog.LevelError)
	case pb.LogLevel_LOG_LEVEL_OFF:
		logLevel.Set(levelOff)
	default:
		logLevel.Set(slog.LevelInfo)
	}
}

// logPanic 把调用边界捕获的 panic 记为一条 ERROR，堆栈放在 stack 字段（CONTRACTS §14.1）；
// 只能在 recover 所在的 defer 中调用，堆栈才包含 panic 的位置。
func logPanic(boundary string, p any) {
	slog.Error(boundary+" panic", "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
}
