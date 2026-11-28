# 未发布

## 新增

- 新增管理端点 `GET /api/v1/core/telemetry?window=5m|15m|30m`，提供宿主机摘要、Core 与各组件进程树的实时 CPU%、常驻内存（RSS）、运行时长、按角色汇总及 30 分钟纯内存历史采样；`GET /api/v1/core` 响应补充 `started_at`。
