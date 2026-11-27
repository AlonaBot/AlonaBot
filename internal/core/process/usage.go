package process

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"time"
)

// TreeUsage 记录单次采样时一个进程或整棵进程树的资源占用快照。
type TreeUsage struct {
	CPUTime      time.Duration // 累计用户态与内核态 CPU 时长；只用于相邻采样窗口做差值。
	RSSBytes     uint64        // 常驻内存字节数（Unix RSS / Windows WorkingSetSize）。
	ProcessCount int           // 归属该树的活跃进程数。
}

// SampleTrees 一次性采集 Core 自身与指定插件进程树的资源快照，并沿用调用方的取消上下文。
// 已正常退出的进程树在返回的 map 中保留零值条目；发生部分读取失败时，将已采集的有效读数与原始错误一起返回。
func SampleTrees(ctx context.Context, processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	return sampleTreesPlatform(ctx, processes)
}

// HostMemoryBytes 返回宿主机物理内存总字节数；当前平台不支持或读取失败时返回 0。
func HostMemoryBytes() uint64 {
	return hostMemoryBytesPlatform()
}

// parseLinuxProcStat 解析 Linux /proc/<pid>/stat 内容，定位最后一个 ')' 以兼容含空格或括号的进程名。
// 返回进程组 ID（pgrp）与 utime+stime 累计时钟滴答数。
func parseLinuxProcStat(raw []byte) (pgrp int, ticks uint64, ok bool) {
	end := bytes.LastIndexByte(raw, ')')
	if end < 0 || end+2 >= len(raw) {
		return 0, 0, false
	}
	// ')' 后的字段索引从 0 开始依次为：0:state, 1:ppid, 2:pgrp, ..., 11:utime, 12:stime。
	fields := strings.Fields(string(raw[end+2:]))
	if len(fields) <= 12 {
		return 0, 0, false
	}
	parsedPgrp, err := strconv.Atoi(fields[2])
	if err != nil {
		return 0, 0, false
	}
	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return parsedPgrp, utime + stime, true
}

// parseLinuxVmRSS 从 /proc/<pid>/status 解析 VmRSS 并转换为字节。
func parseLinuxVmRSS(raw []byte) (uint64, bool) {
	for line := range strings.Lines(string(raw)) {
		value, found := strings.CutPrefix(line, "VmRSS:")
		if !found {
			continue
		}
		fields := strings.Fields(value)
		if len(fields) == 0 {
			return 0, false
		}
		kb, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0, false
		}
		return kb << 10, true
	}
	return 0, false
}

// parseLinuxMemTotal 从 /proc/meminfo 解析 MemTotal 并转换为字节。
func parseLinuxMemTotal(raw []byte) uint64 {
	for line := range strings.Lines(string(raw)) {
		value, found := strings.CutPrefix(line, "MemTotal:")
		if !found {
			continue
		}
		fields := strings.Fields(value)
		if len(fields) == 0 {
			return 0
		}
		kb, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0
		}
		return kb << 10
	}
	return 0
}

// parsePSDuration 解析 macOS/BSD `ps -o time=` 输出的累计 CPU 时长。
// 支持 [[DD-]HH:]MM:SS[.ff] 格式（如 0:01.23、12:34.56、1:02:03、2-03:04:05）。
func parsePSDuration(raw string) (time.Duration, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0, false
	}
	var days uint64
	if dayPart, rest, hasDay := strings.Cut(text, "-"); hasDay {
		parsedDays, err := strconv.ParseUint(dayPart, 10, 64)
		if err != nil {
			return 0, false
		}
		days = parsedDays
		text = rest
	}
	parts := strings.Split(text, ":")
	if len(parts) != 2 && len(parts) != 3 {
		return 0, false
	}
	var hours uint64
	minuteIndex := 0
	if len(parts) == 3 {
		parsedHours, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0, false
		}
		hours = parsedHours
		minuteIndex = 1
	}
	minutes, err := strconv.ParseUint(parts[minuteIndex], 10, 64)
	if err != nil {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(parts[minuteIndex+1], 64)
	if err != nil || seconds < 0 {
		return 0, false
	}
	totalSeconds := float64(days*86400+hours*3600+minutes*60) + seconds
	return time.Duration(totalSeconds * float64(time.Second)), true
}
