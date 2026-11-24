# 未发布

### 修复

- Send/Edit 回执独立返回，AfterSend 按批次排序观察，使用有界接纳及 Core 排空生命周期（#325）。
- 事件期限覆盖排队与 PreEvent，出队过期事件跳过业务处理（#324）。
- Handler 失败保留原始错误及有效 STOP，命令和事件统一 Outcome 归因（#326）。
- 回复修饰改存 Session 隔离的运行期 SQLite 数据库，保留历史编辑，移除随消息增长的内存关联表（#323）。
- plugin、agent 缺省 `sleep_capable` 为 true，构建保留字段省略状态，并说明旧产物的重新构建要求（#322）。
