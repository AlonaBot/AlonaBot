// protocgen 是 contracts 的生成入口：核对 protoc 版本，清除旧生成文件，再用 go.mod 固定的
// protoc-gen-go 重新生成 pb/。由 contracts/doc.go 的 go:generate 在 contracts 目录下调用。
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

// wantProtoc 是生成结果所对应的 protoc 版本；生成文件头会记录它，版本不同会产生无关 diff。
const wantProtoc = "libprotoc 36.0"

const (
	protoDir = "proto"
	outDir   = "pb"
	module   = "github.com/AlonaBot/alonabot/contracts"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "protocgen:", err)
		os.Exit(1)
	}
}

func run() error {
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

	// 删除旧产物，避免已删除的 .proto 留下过期的 .pb.go。
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
