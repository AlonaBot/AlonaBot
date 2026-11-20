# AlonaBot 技术规范

本目录定义现行公开行为、消息模型、协议与管理接口。开发步骤见[源码开发](../docs/development.md)，模块职责和依赖方向见[架构地图](../ARCHITECTURE.md)。

| 文档 | 内容 |
| --- | --- |
| [CONTRACTS-v0.1.md](CONTRACTS-v0.1.md) | API、权限、生命周期、配置、数据、组件安装和运行方式的精确行为 |
| [WIRE-FORMAT.md](WIRE-FORMAT.md) | Protobuf 字段、枚举、控制帧、Bootstrap 与载荷格式 |
| [MESSAGE-MODEL.md](MESSAGE-MODEL.md) | 事件、目标、消息链、发送回执、媒体与平台映射 |
| [ADMIN-API.md](ADMIN-API.md) | 管理认证、组件操作、配置、会话、数据与 WebUI 后台接口 |
| [DESIGN.md](DESIGN.md) | Core、SDK、Adapter 与 Agent 的设计关系及关键取舍 |

## 如何使用

先从开发指南找到操作入口，再按改动涉及的行为阅读对应条款。修改协议字段时同步 `.proto` 源文件、生成产物和协议编号检查；修改 API、配置或生命周期时同步实现、全部调用方、示例与相关验证。

规范中的 D/T/V 标识保留原有追溯身份；公开开发以当前条款和实现为准。平台验收与已知限制见[发布说明](../changelog/v0.1.0/zh-CN.md)，验证方法见[开发指南](../docs/development.md#验证改动)。

## 维护要求

行为条款在对应规范中保持唯一所有者，其他文档通过链接引用。发现实现与规范冲突时，先明确预期行为和影响，由维护者确认后同步修订。保留接口、单位、限额、错误和数据语义；通过测试验证改变，实际命令与结果记录在 PR。
