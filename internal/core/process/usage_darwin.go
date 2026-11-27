//go:build darwin

package process

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// darwinPSPath 固定使用 macOS 系统二进制绝对路径，避免受 PATH 影响。
const (
	darwinPSPath    = "/bin/ps"
	darwinPSTimeout = 3 * time.Second
)

func hostMemoryBytesPlatform() uint64 {
	total, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return 0
	}
	return total
}

func sampleTreesPlatform(processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	byPGID := make(map[int]*Process, len(processes))
	for _, proc := range processes {
		if proc == nil || proc.tree == nil {
			continue
		}
		result[proc] = TreeUsage{}
		byPGID[proc.tree.pgid] = proc
	}
	ctx, cancel := context.WithTimeout(context.Background(), darwinPSTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, darwinPSPath, "-axo", "pid=,pgid=,rss=,time=").Output()
	if err != nil {
		return TreeUsage{}, result, err
	}
	selfPID := os.Getpid()
	var self TreeUsage
	for line := range strings.Lines(string(out)) {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		pgid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		target := byPGID[pgid]
		if pid != selfPID && target == nil {
			continue
		}
		rssKiB, err := strconv.ParseUint(fields[2], 10, 64)
		if err != nil {
			continue
		}
		cpuDuration, ok := parsePSDuration(fields[3])
		if !ok {
			continue
		}
		rssBytes := rssKiB << 10
		if pid == selfPID {
			self = TreeUsage{CPUTime: cpuDuration, RSSBytes: rssBytes, ProcessCount: 1}
		}
		if target != nil {
			usage := result[target]
			usage.CPUTime += cpuDuration
			usage.RSSBytes += rssBytes
			usage.ProcessCount++
			result[target] = usage
		}
	}
	return self, result, nil
}
