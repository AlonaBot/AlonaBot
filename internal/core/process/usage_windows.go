//go:build windows

package process

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// windowsProcessStillActive 对应 Win32 STILL_ACTIVE (259)，用于区分进程已退出与读取失败。
const windowsProcessStillActive = 259

var (
	modKernel32              = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
	procK32GetProcessMemInfo = modKernel32.NewProc("K32GetProcessMemoryInfo")
)

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

func hostMemoryBytesPlatform() uint64 {
	var status memoryStatusEx
	status.Length = uint32(unsafe.Sizeof(status))
	r1, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if r1 == 0 {
		return 0
	}
	return status.TotalPhys
}

func processMemoryBytes(handle windows.Handle) (uint64, error) {
	var counters processMemoryCounters
	counters.CB = uint32(unsafe.Sizeof(counters))
	r1, _, callErr := procK32GetProcessMemInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&counters)), uintptr(counters.CB))
	if r1 == 0 {
		if callErr == nil || callErr == windows.Errno(0) {
			return 0, errors.New("K32GetProcessMemoryInfo failed")
		}
		return 0, callErr
	}
	return uint64(counters.WorkingSetSize), nil
}

func sampleSelfWindows() (TreeUsage, error) {
	handle := windows.CurrentProcess()
	var creation, exit, kernel, user windows.Filetime
	var cpu time.Duration
	var errs []error
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err != nil {
		errs = append(errs, fmt.Errorf("get current process times: %w", err))
	} else {
		ticks := (uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime)) +
			(uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime))
		cpu = time.Duration(ticks) * 100 * time.Nanosecond
	}
	rss, err := processMemoryBytes(handle)
	if err != nil {
		errs = append(errs, fmt.Errorf("read current process memory: %w", err))
	}
	return TreeUsage{
		CPUTime:      cpu,
		RSSBytes:     rss,
		ProcessCount: 1,
	}, errors.Join(errs...)
}

func (t *tree) usage(ctx context.Context) (TreeUsage, error) {
	if err := ctx.Err(); err != nil {
		return TreeUsage{}, err
	}
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.job == 0 {
		// Job 句柄已由 release 关闭，进程树已结束。
		return TreeUsage{}, nil
	}
	var accounting basicAccounting
	if err := windows.QueryInformationJobObject(t.job, windows.JobObjectBasicAccountingInformation,
		uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
		return TreeUsage{}, fmt.Errorf("query job accounting: %w", err)
	}
	cpuTicks := max(int64(0), accounting.TotalUserTime+accounting.TotalKernelTime)
	usage := TreeUsage{
		CPUTime:      time.Duration(cpuTicks) * 100 * time.Nanosecond,
		ProcessCount: int(accounting.ActiveProcesses),
	}
	if accounting.ActiveProcesses == 0 {
		return usage, nil
	}
	// 查询 Job 内当前在途 PID 列表并汇总工作集内存；已读出的 CPU 与进程计数在后续失败时一并保留返回。
	capacity := max(uint32(16), accounting.ActiveProcesses+8)
	headerBytes := unsafe.Sizeof(uint32(0)) * 2
	ptrSize := unsafe.Sizeof(uintptr(0))
	buf := make([]byte, headerBytes+uintptr(capacity)*ptrSize)
	if err := windows.QueryInformationJobObject(t.job, windows.JobObjectBasicProcessIdList,
		uintptr(unsafe.Pointer(&buf[0])), uint32(len(buf)), nil); err != nil {
		return usage, fmt.Errorf("query job process list: %w", err)
	}
	count := *(*uint32)(unsafe.Pointer(&buf[4]))
	limit := min(count, capacity)
	var errs []error
	for i := uint32(0); i < limit; i++ {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			return usage, errors.Join(errs...)
		}
		offset := headerBytes + uintptr(i)*ptrSize
		pid := uint32(*(*uintptr)(unsafe.Pointer(&buf[offset])))
		if pid == 0 {
			continue
		}
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			// Windows 上 PID 在枚举后正常退出时 OpenProcess 返回 ERROR_INVALID_PARAMETER。
			if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
				continue
			}
			errs = append(errs, fmt.Errorf("open process %d: %w", pid, err))
			continue
		}
		rss, memErr := processMemoryBytes(handle)
		if memErr != nil {
			var exitCode uint32
			if exitErr := windows.GetExitCodeProcess(handle, &exitCode); exitErr != nil || exitCode == windowsProcessStillActive {
				errs = append(errs, fmt.Errorf("read process %d memory: %w", pid, memErr))
			}
		} else {
			usage.RSSBytes += rss
		}
		_ = windows.CloseHandle(handle)
	}
	return usage, errors.Join(errs...)
}

func sampleTreesPlatform(ctx context.Context, processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	if err := ctx.Err(); err != nil {
		return TreeUsage{}, result, err
	}
	var errs []error
	self, selfErr := sampleSelfWindows()
	if selfErr != nil {
		errs = append(errs, selfErr)
	}
	for _, proc := range processes {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			return self, result, errors.Join(errs...)
		}
		if proc == nil || proc.tree == nil {
			continue
		}
		usage, err := proc.tree.usage(ctx)
		result[proc] = usage
		if err != nil {
			errs = append(errs, fmt.Errorf("sample process tree %s: %w", proc.PluginID, err))
		}
	}
	return self, result, errors.Join(errs...)
}
