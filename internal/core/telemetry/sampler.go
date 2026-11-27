// Package telemetry 周期性采集 Core 自身与各组件进程树的 CPU、常驻内存与运行时间，
// 并在内存中维护固定容量的环形历史缓冲，供管理 API 只读查询。
//
// 采样不写入 SQLite，零磁盘开销；同进程组件（如纯静态 WebUI 或库模式开发组件）的独立开销记为 0，
// 不重复计入全局总计。CPU 占比统一换算为整机归一化百分比（0..100%）。
//
// 包的组成：
//   - [Sampler] 负责定时采样、CPU 差值归一化与内存环形缓冲；
//   - [Target] 描述一次采样中由调用方提供的组件运行态快照；
//   - [Report] 与 [HistoryPoint] 定义 `GET /api/v1/core/telemetry` 的响应载荷；
//   - [ParseWindow] 校验 `5m` / `15m` / `30m` 历史窗口参数。
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/process"
)

const (
	// DefaultInterval 是后台遥测的默认采样间隔。
	DefaultInterval = 5 * time.Second
	// DefaultHistoryCapacity 保存最近 30 分钟（5 秒间隔 × 360 点）的内存历史样本。
	DefaultHistoryCapacity = 360
	// cpuScaleHostNormalized 标明所有 cpu_percent 字段均按整机核心数归一化到 0..100%。
	cpuScaleHostNormalized = "host_normalized_percent"
)

// Window 表示历史曲线的查询时间窗。
type Window string

const (
	Window5m  Window = "5m"
	Window15m Window = "15m"
	Window30m Window = "30m"
)

// ParseWindow 校验并解析查询窗口；空串使用缺省的 30m。
func ParseWindow(raw string) (Window, error) {
	switch raw {
	case "", string(Window30m):
		return Window30m, nil
	case string(Window5m):
		return Window5m, nil
	case string(Window15m):
		return Window15m, nil
	default:
		return "", fmt.Errorf("invalid window %q: expected 5m, 15m or 30m", raw)
	}
}

func (w Window) duration() time.Duration {
	switch w {
	case Window5m:
		return 5 * time.Minute
	case Window15m:
		return 15 * time.Minute
	default:
		return 30 * time.Minute
	}
}

// Target 描述一个已登记组件在采样时刻的状态与进程句柄。
type Target struct {
	ID        string
	Role      string           // plugin / adapter / agent / webui
	State     string           // 对齐管理 API 的 Package.State（Activated / Suspended / Stopped / Installed / Degraded）
	Running   bool             // 是否处于活跃运行状态
	InProcess bool             // 是否运行于 Core 同进程（或纯静态 WebUI）；为真时不计独立开销
	Process   *process.Process // 独立进程句柄；未运行或同进程时为 nil
	StartedAt time.Time        // 进程启动时刻（UTC）；未运行时为零值
}

// Host 描述当前宿主机的计算资源基线。
type Host struct {
	CPUCores         int    `json:"cpu_cores"`
	MemoryTotalBytes uint64 `json:"memory_total_bytes"`
	Platform         string `json:"platform"`
}

// Aggregate 汇总一组进程的 CPU、常驻内存与活跃进程总数。
type Aggregate struct {
	CPUPercent   float64 `json:"cpu_percent"`
	RSSBytes     uint64  `json:"rss_bytes"`
	ProcessCount int     `json:"process_count"`
}

// CoreProcess 描述 Core 主进程自身的资源占用与运行时间。
type CoreProcess struct {
	PID           int     `json:"pid"`
	StartedAt     string  `json:"started_at"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	CPUPercent    float64 `json:"cpu_percent"`
	RSSBytes      uint64  `json:"rss_bytes"`
	Goroutines    int     `json:"goroutines"`
}

// RoleSummary 按组件角色汇总运行数、总数与资源开销。
type RoleSummary struct {
	CPUPercent   float64 `json:"cpu_percent"`
	RSSBytes     uint64  `json:"rss_bytes"`
	RunningCount int     `json:"running_count"`
	TotalCount   int     `json:"total_count"`
}

// Component 描述单个组件的运行态与资源开销。
type Component struct {
	ID            string  `json:"id"`
	Role          string  `json:"role"`
	State         string  `json:"state"`
	Running       bool    `json:"running"`
	InProcess     bool    `json:"in_process"`
	PID           int     `json:"pid"`
	ProcessCount  int     `json:"process_count"`
	StartedAt     string  `json:"started_at,omitempty"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	CPUPercent    float64 `json:"cpu_percent"`
	RSSBytes      uint64  `json:"rss_bytes"`
}

// HistoryPoint 记录单个采样时刻的总览与各角色内存切片，用于前端趋势图。
type HistoryPoint struct {
	TS              string  `json:"ts"`
	TotalCPUPercent float64 `json:"total_cpu_percent"`
	TotalRSSBytes   uint64  `json:"total_rss_bytes"`
	CoreCPUPercent  float64 `json:"core_cpu_percent"`
	CoreRSSBytes    uint64  `json:"core_rss_bytes"`
	PluginRSSBytes  uint64  `json:"plugin_rss_bytes"`
	AdapterRSSBytes uint64  `json:"adapter_rss_bytes"`
	AgentRSSBytes   uint64  `json:"agent_rss_bytes"`
	WebUIRSSBytes   uint64  `json:"webui_rss_bytes"`
}

// Report 是 GET /api/v1/core/telemetry 的完整响应结构。
type Report struct {
	SampledAt       string                 `json:"sampled_at"`
	IntervalSeconds int                    `json:"interval_seconds"`
	CPUScale        string                 `json:"cpu_scale"`
	Host            Host                   `json:"host"`
	Total           Aggregate              `json:"total"`
	Core            CoreProcess            `json:"core"`
	ByRole          map[string]RoleSummary `json:"by_role"`
	Components      []Component            `json:"components"`
	History         []HistoryPoint         `json:"history"`
}

// Options 配置 [Sampler] 的采样回调与周期参数；未设置的字段使用生产默认值。
type Options struct {
	StartedAt  time.Time
	Targets    func(context.Context) ([]Target, error)
	Sample     func([]*process.Process) (process.TreeUsage, map[*process.Process]process.TreeUsage, error)
	HostMemory func() uint64
	Now        func() time.Time
	Interval   time.Duration
	Capacity   int
	NumCPU     int
	PID        int
}

type cpuSample struct {
	pid int
	cpu time.Duration
	at  time.Time
}

// Sampler 负责定时采集 Core 与组件进程树资源并在内存中保存环形历史。
type Sampler struct {
	startedAt  time.Time
	targets    func(context.Context) ([]Target, error)
	sample     func([]*process.Process) (process.TreeUsage, map[*process.Process]process.TreeUsage, error)
	hostMemory func() uint64
	now        func() time.Time
	interval   time.Duration
	capacity   int
	numCPU     int
	pid        int
	hostMem    uint64

	// sampleMu 串行化采样与前值状态更新，避免并发即时采样与后台定时采样交叉。
	sampleMu   sync.Mutex
	prevSelf   *cpuSample
	prevByComp map[string]cpuSample

	// mu 保护最新快照与环形缓冲的并发读写。
	mu        sync.RWMutex
	hasSample bool
	latest    Report
	ring      []HistoryPoint
	ringHead  int // 下一条写入槽位
	ringCount int // 当前已写入有效点数（<= capacity）
}

// New 创建遥测采样器。
func New(opt Options) *Sampler {
	nowFn := opt.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	startedAt := opt.StartedAt
	if startedAt.IsZero() {
		startedAt = nowFn().UTC()
	}
	interval := opt.Interval
	if interval <= 0 {
		interval = DefaultInterval
	}
	capacity := opt.Capacity
	if capacity <= 0 {
		capacity = DefaultHistoryCapacity
	}
	numCPU := opt.NumCPU
	if numCPU <= 0 {
		numCPU = max(1, runtime.NumCPU())
	}
	pid := opt.PID
	if pid <= 0 {
		pid = os.Getpid()
	}
	sampleFn := opt.Sample
	if sampleFn == nil {
		sampleFn = process.SampleTrees
	}
	hostMemoryFn := opt.HostMemory
	if hostMemoryFn == nil {
		hostMemoryFn = process.HostMemoryBytes
	}
	return &Sampler{
		startedAt:  startedAt.UTC(),
		targets:    opt.Targets,
		sample:     sampleFn,
		hostMemory: hostMemoryFn,
		now:        nowFn,
		interval:   interval,
		capacity:   capacity,
		numCPU:     numCPU,
		pid:        pid,
		hostMem:    hostMemoryFn(),
		prevByComp: make(map[string]cpuSample),
		ring:       make([]HistoryPoint, capacity),
	}
}

// Run 立即执行首次采样并按固定周期持续采样，直到 ctx 取消后退出。
func (s *Sampler) Run(ctx context.Context) {
	if err := s.SampleOnce(ctx); err != nil && ctx.Err() == nil {
		slog.Warn(fmt.Sprintf("Initial telemetry sample failed: %v", err), "error", err)
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.SampleOnce(ctx); err != nil && ctx.Err() == nil {
				slog.Warn(fmt.Sprintf("Periodic telemetry sample failed: %v", err), "error", err)
			}
		}
	}
}

// SampleOnce 执行单次全量采样并更新内存快照与环形历史缓冲。
func (s *Sampler) SampleOnce(ctx context.Context) error {
	s.sampleMu.Lock()
	defer s.sampleMu.Unlock()

	// 1. 拉取当前已登记组件与运行句柄快照，提取独立进程集合。
	var targets []Target
	if s.targets != nil {
		list, err := s.targets(ctx)
		if err != nil {
			return fmt.Errorf("list telemetry targets: %w", err)
		}
		targets = list
	}
	procs := make([]*process.Process, 0, len(targets))
	for _, target := range targets {
		if target.Running && !target.InProcess && target.Process != nil {
			procs = append(procs, target.Process)
		}
	}

	// 2. 一次性采集 Core 自身与全部独立组件进程树的累计 CPU 时长与常驻内存。
	sampledAt := s.now().UTC()
	selfUsage, treeUsages, err := s.sample(procs)
	if err != nil {
		return fmt.Errorf("sample process trees: %w", err)
	}
	if s.hostMem == 0 {
		s.hostMem = s.hostMemory()
	}

	// 3. 计算 Core 自身的整机归一化 CPU% 与运行时长。
	coreCPU := 0.0
	if s.prevSelf != nil && s.prevSelf.pid == s.pid {
		coreCPU = normalizedCPUPercent(selfUsage.CPUTime-s.prevSelf.cpu, sampledAt.Sub(s.prevSelf.at), s.numCPU)
	}
	s.prevSelf = &cpuSample{pid: s.pid, cpu: selfUsage.CPUTime, at: sampledAt}

	coreUptime := uptimeSeconds(sampledAt, s.startedAt)
	coreProcess := CoreProcess{
		PID:           s.pid,
		StartedAt:     s.startedAt.Format(time.RFC3339),
		UptimeSeconds: coreUptime,
		CPUPercent:    roundPercent(coreCPU),
		RSSBytes:      selfUsage.RSSBytes,
		Goroutines:    runtime.NumGoroutine(),
	}

	// 4. 逐组件汇总独立进程开销，按角色归集并计算全局总计。
	byRole := map[string]RoleSummary{
		"plugin":  {},
		"adapter": {},
		"agent":   {},
		"webui":   {},
	}
	totalCPU := coreCPU
	totalRSS := selfUsage.RSSBytes
	totalProcs := max(1, selfUsage.ProcessCount)

	components := make([]Component, 0, len(targets))
	activeIDs := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		summary := byRole[target.Role]
		summary.TotalCount++
		entry := Component{
			ID:        target.ID,
			Role:      target.Role,
			State:     target.State,
			Running:   target.Running,
			InProcess: target.InProcess,
		}
		if target.Running {
			summary.RunningCount++
			if !target.StartedAt.IsZero() {
				entry.StartedAt = target.StartedAt.UTC().Format(time.RFC3339)
				entry.UptimeSeconds = uptimeSeconds(sampledAt, target.StartedAt)
			}
			if !target.InProcess && target.Process != nil {
				pid := target.Process.Pid()
				usage := treeUsages[target.Process]
				compCPU := 0.0
				if prev, ok := s.prevByComp[target.ID]; ok && prev.pid == pid {
					compCPU = normalizedCPUPercent(usage.CPUTime-prev.cpu, sampledAt.Sub(prev.at), s.numCPU)
				}
				s.prevByComp[target.ID] = cpuSample{pid: pid, cpu: usage.CPUTime, at: sampledAt}
				activeIDs[target.ID] = struct{}{}

				entry.PID = pid
				entry.ProcessCount = usage.ProcessCount
				entry.CPUPercent = roundPercent(compCPU)
				entry.RSSBytes = usage.RSSBytes

				summary.CPUPercent += compCPU
				summary.RSSBytes += usage.RSSBytes
				totalCPU += compCPU
				totalRSS += usage.RSSBytes
				totalProcs += usage.ProcessCount
			}
		}
		byRole[target.Role] = summary
		components = append(components, entry)
	}
	// 清理已停止或卸载组件的前次 CPU 样本，避免同名重启误用旧基线。
	for id := range s.prevByComp {
		if _, alive := activeIDs[id]; !alive {
			delete(s.prevByComp, id)
		}
	}
	for role, summary := range byRole {
		summary.CPUPercent = roundPercent(summary.CPUPercent)
		byRole[role] = summary
	}

	point := HistoryPoint{
		TS:              sampledAt.Format(time.RFC3339),
		TotalCPUPercent: roundPercent(totalCPU),
		TotalRSSBytes:   totalRSS,
		CoreCPUPercent:  roundPercent(coreCPU),
		CoreRSSBytes:    selfUsage.RSSBytes,
		PluginRSSBytes:  byRole["plugin"].RSSBytes,
		AdapterRSSBytes: byRole["adapter"].RSSBytes,
		AgentRSSBytes:   byRole["agent"].RSSBytes,
		WebUIRSSBytes:   byRole["webui"].RSSBytes,
	}

	report := Report{
		SampledAt:       sampledAt.Format(time.RFC3339),
		IntervalSeconds: max(1, int(s.interval/time.Second)),
		CPUScale:        cpuScaleHostNormalized,
		Host: Host{
			CPUCores:         s.numCPU,
			MemoryTotalBytes: s.hostMem,
			Platform:         runtime.GOOS + "/" + runtime.GOARCH,
		},
		Total: Aggregate{
			CPUPercent:   roundPercent(totalCPU),
			RSSBytes:     totalRSS,
			ProcessCount: totalProcs,
		},
		Core:       coreProcess,
		ByRole:     byRole,
		Components: components,
	}

	// 5. 发布最新快照并写入固定容量环形历史缓冲。
	s.mu.Lock()
	s.latest = report
	s.hasSample = true
	s.ring[s.ringHead] = point
	s.ringHead = (s.ringHead + 1) % s.capacity
	if s.ringCount < s.capacity {
		s.ringCount++
	}
	s.mu.Unlock()
	return nil
}

// Snapshot 返回最新快照以及指定窗口内按时间正序排列的历史样本切片。
// 若后台尚未完成首采，将同步执行一次即时采样。
func (s *Sampler) Snapshot(ctx context.Context, window Window) (Report, error) {
	s.mu.RLock()
	ready := s.hasSample
	s.mu.RUnlock()
	if !ready {
		if err := s.SampleOnce(ctx); err != nil {
			return Report{}, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	report := s.latest
	report.ByRole = make(map[string]RoleSummary, len(s.latest.ByRole))
	for k, v := range s.latest.ByRole {
		report.ByRole[k] = v
	}
	report.Components = append([]Component(nil), s.latest.Components...)

	maxPoints := max(1, int(window.duration()/s.interval))
	take := min(s.ringCount, maxPoints)
	history := make([]HistoryPoint, take)
	start := (s.ringHead - take + s.capacity) % s.capacity
	for i := 0; i < take; i++ {
		history[i] = s.ring[(start+i)%s.capacity]
	}
	report.History = history
	return report, nil
}

// normalizedCPUPercent 按主机逻辑核心数归一化为 0..100%，负差值（子进程退出回退）钳制为 0。
func normalizedCPUPercent(deltaCPU, deltaWall time.Duration, cores int) float64 {
	if deltaCPU <= 0 || deltaWall <= 0 || cores <= 0 {
		return 0
	}
	percent := float64(deltaCPU) / (float64(deltaWall) * float64(cores)) * 100.0
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func roundPercent(value float64) float64 {
	return math.Round(value*10) / 10
}

func uptimeSeconds(now, startedAt time.Time) int64 {
	if startedAt.IsZero() {
		return 0
	}
	seconds := int64(now.Sub(startedAt).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}
