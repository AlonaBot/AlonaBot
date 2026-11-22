# 整合测试

`agent/` 保留官方 Agent 与真实 Core 的协作断言；`e2e/` 保留 CLI、恢复和跨组件端到端测试。两组测试共用独立的 `go.mod`，组件版本由 `components.lock.json` 固定，Core 使用当前 checkout。Core module 的产品依赖保持独立。

执行命令统一见[开发指南](../docs/development.zh-CN.md#官方组件开发)。

`app/agent_service_test.go.tmpl` 验证 Core 私有装配方法和真实 Agent。运行入口使用临时 modfile 和 Go overlay 把它放回 `internal/app` 测试包；业务断言保留，官方 Agent 资源按锁定工程现场扫描。临时依赖文件由该次运行创建并清理，源码树无需生成测试文件。

发行首次运行检查通过 `ALONABOT_RELEASE_DIR` 消费纯 Core 归档，通过 `ALONABOT_BUNDLE_DIR` 消费内部整合归档。发布工作流同时设置两者，验证空组件清单或 WebUI/Agent 默认状态、登录、重启持久化与正常退出；普通测试不重建发行包。

平台与恢复工具复用 `e2e/` 中的场景，入口见 [scripts/platform](../scripts/platform/README.md) 与 [scripts/scenarios](../scripts/scenarios/README.md)。本目录测试使用独立临时数据和本地模拟服务。
