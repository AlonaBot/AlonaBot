package sdk

import "fmt"

// Generated 提供构建期生成的配置、服务、翻译与 manifest 原文；nil 表示没有对应资源。
type Generated struct {
	ServiceSchemas map[string][]byte // schema_id 对应的规范 JSON Service Schema。
	I18n           map[string][]byte // i18n 文件名对应的原文。
	ConfigSchema   []byte            // config_schema.json 的规范 JSON。
	Manifest       []byte            // 入口包 plugin.yaml 原文。
}

// UseGenerated 把只读构建资源显式绑定到插件对象；重复绑定在 Collect 时报错。
func UseGenerated(p *Plugin, g Generated) {
	p.checkRegistration()
	p.generatedCount++
	if p.generatedCount > 1 {
		p.errs = append(p.errs, fmt.Errorf("generated resources bound more than once"))
	}
	p.generated = g
}

// Generated 返回当前插件绑定的只读构建资源。
func (p *Plugin) Generated() Generated { return p.generated }

func generatedFor(p *Plugin) Generated { return p.generated }
