package main

import (
	"errors"
	"os"
)

// shortcut: Windows 原生观测尚待接入，实际运行明确报告未实现。
func processExists(int) (bool, error) {
	return false, errors.New("Windows process inspection is not implemented")
}

func processGroup(int) []int { return nil }

func lockHeld(string) (bool, error) {
	return false, errors.New("Windows lock inspection is not implemented")
}

func killedState(*os.ProcessState) bool { return false }

func addressInUse(error) bool { return false }
