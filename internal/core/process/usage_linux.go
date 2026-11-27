//go:build linux

package process

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

const (
	// linuxAuxvClkTck 是 ELF 辅助向量中 AT_CLKTCK 的固定键，用于读取系统每秒时钟滴答数。
	linuxAuxvClkTck = 17
	// defaultLinuxClkTck 是绝大多数 Linux 发行版的默认 USER_HZ。
	defaultLinuxClkTck = 100
)

var (
	clkTckOnce  sync.Once
	linuxClkTck uint64 = defaultLinuxClkTck
)

func readLinuxClkTck() uint64 {
	clkTckOnce.Do(func() {
		raw, err := os.ReadFile("/proc/self/auxv")
		if err != nil {
			return
		}
		word := int(unsafe.Sizeof(uintptr(0)))
		entry := word * 2
		for offset := 0; offset+entry <= len(raw); offset += entry {
			var key, value uint64
			if word == 8 {
				key = binary.NativeEndian.Uint64(raw[offset : offset+8])
				value = binary.NativeEndian.Uint64(raw[offset+8 : offset+16])
			} else {
				key = uint64(binary.NativeEndian.Uint32(raw[offset : offset+4]))
				value = uint64(binary.NativeEndian.Uint32(raw[offset+4 : offset+8]))
			}
			if key == 0 {
				break
			}
			if key == linuxAuxvClkTck && value > 0 {
				linuxClkTck = value
				return
			}
		}
	})
	return linuxClkTck
}

func hostMemoryBytesPlatform() uint64 {
	raw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	return parseLinuxMemTotal(raw)
}

func sampleTreesPlatform(ctx context.Context, processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	byPGID := make(map[int]*Process, len(processes))
	for _, proc := range processes {
		if proc == nil || proc.tree == nil {
			continue
		}
		result[proc] = TreeUsage{}
		byPGID[proc.tree.pgid] = proc
	}
	if err := ctx.Err(); err != nil {
		return TreeUsage{}, result, err
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return TreeUsage{}, result, err
	}
	selfPID := os.Getpid()
	hz := readLinuxClkTck()
	var self TreeUsage
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return self, result, err
		}
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		statRaw, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if err != nil {
			continue // 进程在两次读取之间退出时跳过。
		}
		pgrp, ticks, ok := parseLinuxProcStat(statRaw)
		if !ok {
			continue
		}
		target := byPGID[pgrp]
		if pid != selfPID && target == nil {
			continue
		}
		statusRaw, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "status"))
		if err != nil {
			continue
		}
		rss, _ := parseLinuxVmRSS(statusRaw)
		cpuDuration := time.Duration(ticks) * time.Second / time.Duration(hz)
		if pid == selfPID {
			self = TreeUsage{CPUTime: cpuDuration, RSSBytes: rss, ProcessCount: 1}
		}
		if target != nil {
			usage := result[target]
			usage.CPUTime += cpuDuration
			usage.RSSBytes += rss
			usage.ProcessCount++
			result[target] = usage
		}
	}
	return self, result, nil
}
