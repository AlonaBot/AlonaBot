//go:build unix

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	componentDirectories = map[string]string{}
	componentLock        map[string]struct{ Module, Version string }
)

// loadComponentDirectories 使用发行锁的精确版本，避免整合测试隐式读取邻接 checkout。
func loadComponentDirectories() {
	raw, err := os.ReadFile("../../components.lock.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(raw, &componentLock); err != nil {
		panic(err)
	}
	for name, source := range componentLock {
		output, err := exec.Command("go", "mod", "download", "-json", source.Module+"@"+source.Version).CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("download %s: %v\n%s", name, err, output))
		}
		var downloaded struct{ Dir string }
		if err := json.Unmarshal(output, &downloaded); err != nil {
			panic(err)
		}
		directory := filepath.Join(binDir, "sources", name)
		if err := os.CopyFS(directory, os.DirFS(downloaded.Dir)); err != nil {
			panic(err)
		}
		componentDirectories[name] = directory
	}
}

// CLI 的公开退出码与临时终端实例身份由端到端断言直接核对。
const (
	exitFailure      = 1
	exitUsage        = 2
	consoleAdapterID = "console"
)

type sleepPlugin struct {
	ID      string          `json:"id"`
	Policy  json.RawMessage `json:"policy"`
	Runtime json.RawMessage `json:"runtime"`
}
