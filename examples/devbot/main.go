// devbot 在当前进程运行一个插件和回环 Adapter，展示信号驱动的库模式生命周期。
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlonaBot/AlonaBot"
	"github.com/AlonaBot/AlonaBot/examples/devbot/adapter"
	"github.com/AlonaBot/AlonaBot/examples/devbot/plugin"
)

func main() {
	dir := flag.String("data-dir", "data", "Core data directory")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	core, err := alonabot.Start(ctx, alonabot.DataDir(*dir), alonabot.Plugin(plugin.New()), alonabot.Adapter(adapter.New()))
	if err != nil {
		slog.Error("devbot startup failed", "error", err)
		os.Exit(1)
	}
	slog.Info("devbot ready")
	if err := core.Wait(); err != nil {
		slog.Error("devbot stopped with an error", "error", err)
		os.Exit(1)
	}
}
