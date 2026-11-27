//go:build windows

package process

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

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

func processMemoryBytes(handle windows.Handle) uint64 {
	var counters processMemoryCounters
	counters.CB = uint32(unsafe.Sizeof(counters))
	r1, _, _ := procK32GetProcessMemInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&counters)), uintptr(counters.CB))
	if r1 == 0 {
		return 0
	}
	return uint64(counters.WorkingSetSize)
}

func sampleSelfWindows() TreeUsage {
	handle := windows.CurrentProcess()
	var creation, exit, kernel, user windows.Filetime
	var cpu time.Duration
	if err := windows.GetProcessTimes(handle, &creation, &exit, &kernel, &user); err == nil {
		ticks := (uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime)) +
			(uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime))
		cpu = time.Duration(ticks) * 100 * time.Nanosecond
	}
	return TreeUsage{
		CPUTime:      cpu,
		RSSBytes:     processMemoryBytes(handle),
		ProcessCount: 1,
	}
}

func (t *tree) usage() TreeUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.job == 0 {
		return TreeUsage{}
	}
	var accounting basicAccounting
	if err := windows.QueryInformationJobObject(t.job, windows.JobObjectBasicAccountingInformation,
		uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
		return TreeUsage{}
	}
	cpuTicks := max(int64(0), accounting.TotalUserTime+accounting.TotalKernelTime)
	usage := TreeUsage{
		CPUTime:      time.Duration(cpuTicks) * 100 * time.Nanosecond,
		ProcessCount: int(accounting.ActiveProcesses),
	}
	// 查询 Job 内当前在途 PID 列表并汇总工作集内存。
	capacity := max(uint32(16), accounting.ActiveProcesses+8)
	headerBytes := unsafe.Sizeof(uint32(0)) * 2
	ptrSize := unsafe.Sizeof(uintptr(0))
	buf := make([]byte, headerBytes+uintptr(capacity)*ptrSize)
	if err := windows.QueryInformationJobObject(t.job, windows.JobObjectBasicProcessIdList,
		uintptr(unsafe.Pointer(&buf[0])), uint32(len(buf)), nil); err != nil {
		return usage
	}
	count := *(*uint32)(unsafe.Pointer(&buf[4]))
	limit := min(count, capacity)
	for i := uint32(0); i < limit; i++ {
		offset := headerBytes + uintptr(i)*ptrSize
		pid := uint32(*(*uintptr)(unsafe.Pointer(&buf[offset])))
		if pid == 0 {
			continue
		}
		handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			continue
		}
		usage.RSSBytes += processMemoryBytes(handle)
		_ = windows.CloseHandle(handle)
	}
	return usage
}

func sampleTreesPlatform(processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	for _, proc := range processes {
		if proc == nil || proc.tree == nil {
			continue
		}
		result[proc] = proc.tree.usage()
	}
	return sampleSelfWindows(), result, nil
}
