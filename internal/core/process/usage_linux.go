//go:build linux

package process

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// linuxAuxvClkTck 对应 ELF 辅助向量中的 AT_CLKTCK。
const linuxAuxvClkTck = 17

// 系统时钟频率在进程生命周期中不变，缓存实际值与读取错误。
var linuxClockTicks = sync.OnceValues(func() (uint64, error) {
	raw, err := os.ReadFile("/proc/self/auxv")
	if err != nil {
		return 0, fmt.Errorf("read clock ticks: %w", err)
	}
	word := int(unsafe.Sizeof(uintptr(0)))
	for offset := 0; offset+word*2 <= len(raw); offset += word * 2 {
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
			return value, nil
		}
	}
	return 0, fmt.Errorf("AT_CLKTCK is missing from /proc/self/auxv")
})

func hostMemoryBytesPlatform() (uint64, error) {
	raw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
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
	hz, clockErr := linuxClockTicks()
	var errs []error
	if clockErr != nil {
		errs = append(errs, clockErr)
	}
	var self TreeUsage
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return self, result, errors.Join(append(errs, err)...)
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
			// 退出与其他读取错误分别处理；未知归属的系统进程留给下一次枚举。
			if (pid == selfPID || byPGID[pid] != nil) && !errors.Is(err, os.ErrNotExist) {
				errs = append(errs, fmt.Errorf("read process %d stat: %w", pid, err))
			}
			continue
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
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errs = append(errs, fmt.Errorf("read process %d status: %w", pid, err))
		}
		rss, _ := parseLinuxVmRSS(statusRaw)
		var cpuDuration time.Duration
		if clockErr == nil {
			// 先拆分整秒，避免长时间运行进程的累计 ticks 在乘纳秒时溢出。
			cpuDuration = time.Duration(ticks/hz)*time.Second + time.Duration(ticks%hz)*time.Second/time.Duration(hz)
		}
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
	return self, result, errors.Join(errs...)
}
