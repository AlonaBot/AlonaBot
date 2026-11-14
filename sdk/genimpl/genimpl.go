// Package genimpl 提供 CLI 注入的登记代码所需的资源入口。
//
// 本包只供构建工具生成的代码使用，不对组件作者承诺兼容。
package genimpl

import "github.com/AlonaBot/AlonaBot/internal/generated"

// Resources 保存构建工具扫描得到的只读资源。
type Resources = generated.Resources

// Register 按入口包登记资源；重复登记同一路径会 panic。
func Register(pkgPath string, r Resources) { generated.Register(pkgPath, r) }
