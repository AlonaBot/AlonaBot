// 维护者：FlanChanXwO
// 创建日期：2026-10-21
// 职责：contracts 包的包文档与 go:generate 入口。

// Package contracts 存放协议 1.0 的 Protobuf 源文件（proto/）、生成类型（pb/）和纯函数校验。
// 本包及子包不依赖本 module 的其他包，也不做运行时 I/O。
//
// 在仓库根执行 go generate ./... 重新生成，不要手改 pb/ 下的文件。
//
// testdata 中的机器格式文件：frames/*.bin 是 golden 帧，internal/wire 的集成测试会读取，
// 可用本地测试 TestGoldenFrames -update 重写；manifest/*.yaml 与 release/release.json 供本地测试使用，
// release.json 中的 publisher_key_id 是占位值，测试时替换为测试公钥的 KeyID 后再签名。
package contracts

//go:generate go run ./internal/protocgen
