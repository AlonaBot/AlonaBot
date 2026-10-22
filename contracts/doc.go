// Package contracts 存放 AlonaBot 协议 1.0 的 Protobuf 源文件（proto/）、生成类型（pb/）
// 与纯函数校验。本包及其子包不依赖本 module 的其他包，也不做运行时 I/O。
//
// 生成：在仓库根执行 go generate ./...；不要手改 pb/ 下的文件。
//
// testdata 中不能写注释的机器格式：frames/*.bin 是 validate 测试的 golden 帧，由
// go test ./contracts/validate -run TestGoldenFrames -update 重写；release/release.json 是 release
// 验签样例，其中 publisher_key_id 为占位值，测试时替换为测试公钥的 KeyID 后再签名。
package contracts

//go:generate go run ./internal/protocgen
