//go:build unix

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func prepareRuntime(*exec.Cmd) {}

func interruptRuntime(process *os.Process) error { return process.Signal(os.Interrupt) }

func checkTokenPermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != fileMode {
		return fmt.Errorf("first token mode = %s; want 0600", info.Mode())
	}
	return nil
}
