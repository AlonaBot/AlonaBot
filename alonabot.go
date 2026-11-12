// Package alonabot 在当前进程装配 Core，并通过 loopback Session 运行开发组件。
//
// Start 等全部组件激活后返回，调用方取消 context 后通过 Core.Wait 等待停止。
// Run 同时处理 SIGINT 与 SIGTERM。已安装组件仍由独立进程运行。
package alonabot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/adapter"
)

// Option 配置一次 Core 启动。
type Option func(*options)

type options struct {
	dir        string
	args       []string
	components []component
}

// Plugin 登记由生成资源确定身份的开发插件。
func Plugin(root sdk.Root) Option {
	return func(o *options) { o.components = append(o.components, component{root: root, role: state.RolePlugin}) }
}

// Adapter 登记由生成资源确定身份的开发 Adapter。
func Adapter(a adapter.Adapter) Option {
	return func(o *options) { o.components = append(o.components, component{adapter: a, role: state.RoleAdapter}) }
}

// DataDir 指定 Core 数据目录，缺省为当前工作目录下的 data。
func DataDir(dir string) Option { return func(o *options) { o.dir = dir } }

// Args 追加调优参数或参数文件，后于 ALONABOT_ARGS 应用。
func Args(values ...string) Option {
	args := append([]string(nil), values...)
	return func(o *options) { o.args = append(o.args, args...) }
}

// Core 持有已装配的运行资源；取消 Start 的 context 后调用 Wait 等待停止结果。
type Core struct {
	assembly   *assembly
	store      *state.Store
	output     *logging.Output
	values     tune.Values
	components []*runningComponent
	done       chan struct{} // err 写入后关闭，允许多个 Wait 读取同一结果。
	err        error
}

// Run 启动 Core，并等待 SIGINT、SIGTERM 或组件异常后完成停止。
func Run(opts ...Option) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	core, err := Start(ctx, opts...)
	if err != nil {
		return err
	}
	return core.Wait()
}

// Start 完成身份检查、握手、注册与激活后返回 Core；失败时清理已启动的部分。
func Start(ctx context.Context, opts ...Option) (*Core, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o := options{dir: "data"}
	for _, apply := range opts {
		apply(&o)
	}
	args := o.args
	if env := os.Getenv("ALONABOT_ARGS"); env != "" {
		args = append([]string{env}, args...)
	}
	values, overrides, err := tune.Parse(args)
	if err != nil {
		return nil, err
	}
	dir, err := filepath.Abs(o.dir)
	if err != nil {
		return nil, err
	}
	// 1. 取得实例锁后才打开日志和媒体目录，避免改动另一个 Core 的资源。
	store, err := state.Open(ctx, dir)
	if err != nil {
		return nil, err
	}
	c := &Core{store: store, values: values, done: make(chan struct{})}
	fail := func(err error) (*Core, error) { return nil, errors.Join(err, c.shutdown()) }
	installs, err := store.Installs(ctx)
	if err != nil {
		return fail(err)
	}
	if err := validateComponents(o.components, installs); err != nil {
		return fail(err)
	}
	output, err := logging.Open(logging.Options{Format: logging.FormatText, Level: slog.LevelInfo, Stderr: os.Stderr})
	if err != nil {
		return fail(err)
	}
	c.output = output
	if err := output.EnableFile(logging.FileOptions{Dir: filepath.Join(dir, "logs"), MaxSize: logging.DefaultMaxFileSize, MaxTotal: logging.DefaultMaxTotalSize, MaxAgeDays: logging.DefaultMaxAgeDays}); err != nil {
		return fail(err)
	}
	slog.SetDefault(output.Logger())
	for _, override := range overrides {
		slog.Warn("tuning parameter overridden", "name", override.Name, "value", override.Value, "origin", override.Origin)
	}
	if values.CoreMemorySoftLimit > 0 {
		debug.SetMemoryLimit(values.CoreMemorySoftLimit)
	}
	// 2. 装配公共服务，再依次启动已启用安装包与显式登记的开发组件。
	c.assembly, err = assemble(store, dir, values)
	if err != nil {
		return fail(err)
	}
	for _, install := range installs {
		if install.Source == devSource {
			continue
		}
		enabled, err := store.Enabled(ctx, install.PluginID)
		if err != nil {
			return fail(err)
		}
		if enabled {
			if err := c.startInstalled(ctx, dir, install); err != nil {
				return fail(fmt.Errorf("start %s: %w", install.PluginID, err))
			}
		}
	}
	for i := range o.components {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if err := c.startDevelopment(ctx, dir, &o.components[i]); err != nil {
			return fail(fmt.Errorf("start %s: %w", o.components[i].manifest.ID, err))
		}
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	c.assembly.router.Start()
	// 3. 一个所有者处理取消与异常，全部组件共用一次停止起点。
	go c.run(ctx)
	return c, nil
}

// Wait 等待 Core 停止；组件超时或异常退出时返回含组件 ID 的错误。
func (c *Core) Wait() error { <-c.done; return c.err }

func (c *Core) run(ctx context.Context) {
	failed := make(chan error, len(c.components))
	for _, running := range c.components {
		go func() {
			select {
			case <-running.exited():
			case <-running.instance.Session.Done():
			case <-c.done:
				return
			}
			failed <- fmt.Errorf("component %s exited before Core shutdown", running.id)
		}()
	}
	ticker := time.NewTicker(c.values.MediaTTL)
	defer ticker.Stop()
running:
	for {
		select {
		case <-ctx.Done():
			break running
		case err := <-failed:
			c.err = err
			break running
		case <-ticker.C:
			c.assembly.media.Sweep()
		}
	}
	c.err = errors.Join(c.err, c.shutdown())
	close(c.done)
}

func (c *Core) shutdown() error {
	t0 := time.Now()
	timing := host.NewStopTiming(c.values)
	var wg sync.WaitGroup
	errs := make(chan error, len(c.components)+1)
	// Router 的排空与组件冻结并行，避免把停止预算逐个累加。
	if c.assembly != nil {
		wg.Go(func() {
			ctx, cancel := context.WithDeadline(context.Background(), t0.Add(timing.Drain))
			defer cancel()
			if err := c.assembly.router.Close(ctx); err != nil {
				slog.Warn("event router drain ended", "error", err)
			}
		})
	}
	for _, running := range c.components {
		wg.Go(func() {
			if err := c.stopComponent(running, t0, timing); err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(errs)
	var result error
	for err := range errs {
		result = errors.Join(result, err)
	}
	if c.assembly != nil {
		c.assembly.scheduler.Close()
	}
	// 同进程用户回调超时时保留实例锁，实际退出后再关闭状态库与日志。
	pending := false
	for _, running := range c.components {
		select {
		case <-running.exited():
		default:
			pending = true
		}
	}
	if pending {
		go func() {
			for _, running := range c.components {
				<-running.exited()
			}
			if err := c.close(); err != nil {
				slog.Error("close Core resources failed", "error", err)
			}
		}()
		return result
	}
	return errors.Join(result, c.close())
}

func (c *Core) close() error {
	err := c.store.Close()
	if c.output != nil {
		err = errors.Join(err, c.output.Close())
	}
	return err
}
