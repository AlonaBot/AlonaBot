//go:build windows

package process

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// tree 是由 Core 独占持有的 Job Object；设置 KILL_ON_JOB_CLOSE，句柄关闭时终止其中所有进程。
// shortcut: 进程在 Start 之后才加入 Job，加入前创建的孙进程不在管理范围；需要时改为挂起创建再加入。
type tree struct{ job windows.Handle }

func prepareTree(cmd *exec.Cmd) {}

func attachTree(cmd *exec.Cmd) (*tree, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	h, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	defer windows.CloseHandle(h)
	if err := windows.AssignProcessToJobObject(job, h); err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	return &tree{job: job}, nil
}

func (t *tree) kill() error { return windows.TerminateJobObject(t.job, 1) }

// basicAccounting 对应 Win32 JOBOBJECT_BASIC_ACCOUNTING_INFORMATION（x/sys/windows 未提供该结构）。
type basicAccounting struct {
	// TotalUserTime 等四项是 Job 的 CPU 时间统计（100ns 单位），此处只为保持内存布局。
	TotalUserTime, TotalKernelTime, ThisPeriodTotalUserTime, ThisPeriodTotalKernelTime int64
	// TotalPageFaultCount 是 Job 内进程的累计缺页数，此处不用。
	TotalPageFaultCount uint32
	// TotalProcesses 是曾加入 Job 的进程总数，此处不用。
	TotalProcesses uint32
	// ActiveProcesses 是 Job 内仍在运行的进程数；为 0 表示进程树已全部退出。
	ActiveProcesses uint32
	// TotalTerminatedProcesses 是因违反 Job 限制而被终止的进程数，此处不用。
	TotalTerminatedProcesses uint32
}

// empty 报告 Job 内是否已没有仍在运行的进程。
func (t *tree) empty() (bool, error) {
	var info basicAccounting
	if err := windows.QueryInformationJobObject(t.job, windows.JobObjectBasicAccountingInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)), nil); err != nil {
		return false, err
	}
	return info.ActiveProcesses == 0, nil
}

func (t *tree) release() error { return windows.CloseHandle(t.job) }
