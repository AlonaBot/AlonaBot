//go:build unix

package main

import (
	"errors"
	"os"
	"syscall"
)

func processExists(pid int) (bool, error) {
	err := syscall.Kill(pid, 0)
	if errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	if errors.Is(err, syscall.EPERM) {
		return true, nil
	}
	return err == nil, err
}

func processGroup(pid int) []int { return []int{-pid} }

func lockHeld(path string) (held bool, err error) {
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if !errors.Is(err, syscall.EINTR) {
			break
		}
	}
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func killedState(state *os.ProcessState) bool {
	status := state.Sys().(syscall.WaitStatus)
	return status.Signaled() && status.Signal() == syscall.SIGKILL
}

func addressInUse(err error) bool { return errors.Is(err, syscall.EADDRINUSE) }
