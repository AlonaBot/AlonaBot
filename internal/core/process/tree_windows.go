//go:build windows

package process

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// tree 持有一个由 Core 独占的 Job Object；设置了 KILL_ON_JOB_CLOSE，句柄关闭时终止其中所有进程。
// shortcut: 进程在 Start 之后才加入 Job，加入前创建的孙进程不在管理范围；需要时改为挂起创建再加入。
type tree struct {
	// Core 独占持有的 Job Object 句柄，release 时关闭。
	job windows.Handle
}

// prepareTree 无需在启动前设置：进程在启动后才加入 Job（见类型上的 shortcut）。
func prepareTree(cmd *exec.Cmd) {}

// attachTree 创建带 KILL_ON_JOB_CLOSE 的 Job Object，并把已启动的子进程加入其中。
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

// kill 终止 Job 内的全部进程；不等待它们退出。
func (t *tree) kill() error { return windows.TerminateJobObject(t.job, 1) }

// basicAccounting 对应 Win32 JOBOBJECT_BASIC_ACCOUNTING_INFORMATION（x/sys/windows 未提供该结构）。
type basicAccounting struct {
	TotalUserTime             int64  // 累计用户态 CPU 时间（100ns 单位），此处只为保持内存布局
	TotalKernelTime           int64  // 累计内核态 CPU 时间（100ns 单位），此处不用
	ThisPeriodTotalUserTime   int64  // 本计费周期的用户态 CPU 时间，此处不用
	ThisPeriodTotalKernelTime int64  // 本计费周期的内核态 CPU 时间，此处不用
	TotalPageFaultCount       uint32 // Job 内进程的累计缺页数，此处不用
	TotalProcesses            uint32 // 曾加入 Job 的进程总数，此处不用
	ActiveProcesses           uint32 // Job 内仍在运行的进程数；为 0 表示进程树已全部退出
	TotalTerminatedProcesses  uint32 // 因违反 Job 限制而被终止的进程数，此处不用
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

// reap 无需回收：Windows 没有僵尸进程，退出的进程不再计入 ActiveProcesses。
func (t *tree) reap() {}

// release 关闭 Job 句柄；KILL_ON_JOB_CLOSE 使仍在 Job 中的进程随之终止。重复调用无副作用。
func (t *tree) release() error {
	if t.job == 0 {
		return nil
	}
	err := windows.CloseHandle(t.job)
	t.job = 0
	return err
}
