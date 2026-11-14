// Package llmagent 为插件测试套件提供只登记文本生成服务的开发 Agent。
//
// 请求与响应由共享的 llm Schema 校验，业务结果由每次构造传入的函数决定。
package llmagent

import (
	"context"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
)

// New 按生成器的工程入口约定构造并绑定资源；服务由 Handle 登记。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()

	return p
}

// Handle 登记生成服务，将本次 Service 的上下文与请求交给 handler。
func Handle(p *sdk.Plugin, handler func(context.Context, llm.Request) (llm.Response, error)) {
	sdk.Serve(p, llm.GenerateService).Access(sdk.AccessPublic).Handle(func(c *sdk.Context, req llm.Request) (llm.Response, error) {
		return handler(c.Context(), req)
	})
}
