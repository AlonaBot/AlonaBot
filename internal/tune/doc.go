// Package tune 管理启动时的调优参数表、生效值与覆盖来源。
//
// 参数名称、单位、缺省值和约束遵循 CONTRACTS §16。
// [Parse] 解析直接参数或文本文件，[Defaults] 提供缺省值，[WriteList] 列出全部参数，[Values] 供调用方装配运行时。
package tune
