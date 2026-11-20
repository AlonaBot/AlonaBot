package main

import (
	"context"
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

// shortcut: Windows 恢复需原生进程与检查点驱动，后续替换此入口。
func (*report) runScenarios(context.Context, options) error {
	return errors.New("Windows native recovery scenarios are not implemented")
}
