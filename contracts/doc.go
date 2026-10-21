// Package contracts 存放 AlonaBot 协议 1.0 的 Protobuf 源文件（proto/）、生成类型（pb/）
// 与纯函数校验。本包及其子包不依赖本 module 的其他包，也不做运行时 I/O。
//
// 生成：在仓库根执行 go generate ./...；不要手改 pb/ 下的文件。
package contracts

//go:generate go run ./internal/protocgen
