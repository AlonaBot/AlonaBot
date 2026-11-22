// Package contracts 存放 AlonaBot 协议 1.0 的 Protobuf 源文件、生成类型和纯函数校验。
//
// proto/ 下是协议源文件，pb/ 下是生成的 Go 类型，子包 validate、manifest、release 提供
// 不依赖运行时状态的校验。本包及其子包不依赖本 module 的其他包，也不做运行时 I/O。
//
// 修改协议时只改 proto/ 下的源文件，然后在仓库根执行 go generate ./... 重新生成；不要手改 pb/。
// 字段号与枚举值只能追加，删除时必须 reserved。
//
// testdata/frames/*.bin 是 golden 帧，internal/wire 的集成测试读取它们；
// testdata/manifest 供本地测试使用。
package contracts

//go:generate go run ./internal/protocgen
