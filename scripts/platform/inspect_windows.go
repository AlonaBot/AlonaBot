package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	processSynchronize                   = 0x00100000
	errorInvalidParameter  syscall.Errno = 87
	errorLockViolation     syscall.Errno = 33
	wsaAddressInUse        syscall.Errno = 10048
	lockExclusiveImmediate               = 0x00000002 | 0x00000001
	forceKillExitCode                    = 1 // os.Process.Kill 在 Windows 上调用 TerminateProcess(handle, 1)。
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	generateConsoleCtrlEvent = kernel32.NewProc("GenerateConsoleCtrlEvent")
	lockFileEx               = kernel32.NewProc("LockFileEx")
	unlockFileEx             = kernel32.NewProc("UnlockFileEx")
)

func prepareCore(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func interruptCore(process *os.Process) error {
	// CTRL_BREAK_EVENT 定向发送给本次 Core 的进程组，Core 的 signal.NotifyContext 负责正常退出。
	ok, _, err := generateConsoleCtrlEvent.Call(syscall.CTRL_BREAK_EVENT, uintptr(process.Pid))
	if ok == 0 {
		return err
	}
	return nil
}

func processExists(pid int) (exists bool, err error) {
	handle, err := syscall.OpenProcess(processSynchronize, false, uint32(pid))
	if errors.Is(err, errorInvalidParameter) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, syscall.CloseHandle(handle)) }()
	status, err := syscall.WaitForSingleObject(handle, 0)
	return status == syscall.WAIT_TIMEOUT, err
}

func processGroup(int) []int { return nil }

func lockHeld(path string) (held bool, err error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	// 与 internal/oslock 相同，非阻塞独占文件起始的一个字节。
	overlapped := new(syscall.Overlapped)
	ok, _, lockErr := lockFileEx.Call(file.Fd(), lockExclusiveImmediate, 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
	if ok == 0 {
		if errors.Is(lockErr, errorLockViolation) {
			return true, nil
		}
		return false, lockErr
	}
	ok, _, err = unlockFileEx.Call(file.Fd(), 0, 1, 0, uintptr(unsafe.Pointer(overlapped)))
	if ok == 0 {
		return false, err
	}
	return false, nil
}

func killedState(state *os.ProcessState) bool { return state.ExitCode() == forceKillExitCode }

func addressInUse(err error) bool { return errors.Is(err, wsaAddressInUse) }
