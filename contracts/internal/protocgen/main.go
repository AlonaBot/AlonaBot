// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：生成 contracts/pb：核对 protoc 版本，删除旧的生成文件，再用 go.mod 固定的 protoc-gen-go 重新生成。由 contracts/doc.go 的 go:generate 调用。

package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// wantProtoc 固定生成所用的 protoc 版本；生成文件头会记录它，版本不同会产生无关 diff。
const wantProtoc = "libprotoc 36.0"

const (
	protoDir = "proto"
	outDir   = "pb"
	module   = "github.com/AlonaBot/AlonaBot/contracts"
)

// main 运行生成流程，失败时打印原因并以非零状态退出。
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "protocgen:", err)
		os.Exit(1)
	}
}

// run 核对工具版本、收集 .proto、清除旧产物，然后调用 protoc 重新生成 pb/。
func run() error {
	// 1. 核对 protoc 版本并定位 go.mod 固定的 protoc-gen-go。
	ver, err := output("protoc", "--version")
	if err != nil {
		return fmt.Errorf("执行 protoc --version 失败: %w", err)
	}
	if ver != wantProtoc {
		return fmt.Errorf("protoc 版本为 %q，需要 %q", ver, wantProtoc)
	}
	plugin, err := output("go", "tool", "-n", "protoc-gen-go")
	if err != nil {
		return fmt.Errorf("定位 protoc-gen-go 失败: %w", err)
	}

	// 2. 按路径排序收集全部 .proto，保证生成顺序稳定。
	var protos []string
	err = filepath.WalkDir(protoDir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".proto") {
			rel, _ := filepath.Rel(protoDir, p)
			protos = append(protos, filepath.ToSlash(rel))
		}
		return err
	})
	if err != nil {
		return err
	}
	if len(protos) == 0 {
		return fmt.Errorf("%s 下没有 .proto 文件", protoDir)
	}
	sort.Strings(protos)

	// 3. 删除旧产物，避免已删除的 .proto 留下过期的 .pb.go。
	err = filepath.WalkDir(outDir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".pb.go") {
			return os.Remove(p)
		}
		if os.IsNotExist(err) {
			return nil
		}
		return err
	})
	if err != nil {
		return err
	}

	// 4. 以 module 选项生成，输出路径与 go_package 对应。
	args := append([]string{
		"--plugin=protoc-gen-go=" + plugin,
		"-I", protoDir,
		"--go_out=.",
		"--go_opt=module=" + module,
	}, protos...)
	cmd := exec.Command("protoc", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// output 运行命令并返回去掉首尾空白的标准输出；失败时错误中附带 stderr。
func output(name string, args ...string) (string, error) {
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
