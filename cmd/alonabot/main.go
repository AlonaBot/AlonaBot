// alonabot 命令通过根包装配 Core，经管理 HTTP API 操作组件。
package main

import "os"

const (
	exitFailure = 1 // 运行失败。
	exitUsage   = 2 // 参数或设置错误。
)

func main() { os.Exit(execute(os.Args[1:])) }
