# 更新日志

每个版本目录包含 `en.md` 与 `zh-CN.md`。日常变更写入 `unreleased/`；发行准备时整理到对应版本目录。条目说明用户可见的变化，按 Added、Changed、Deprecated、Removed、Fixed、Security 分类，省略空分类，相关 PR 链接放在条目旁。

每条用完整句先写用户或开发者现在能做什么、行为有何变化。能帮助读者定位时再写 API 名称；实现细节放在关联 PR。确实要改命令、配置或组件时，在对应条目写清谁需要操作、怎么做。常规部署步骤留在部署指南；无需操作的变更直接写结果。

每份 `v<SemVer>/{en,zh-CN}.md` 采用 Keep a Changelog 单版本格式，以 `## [<SemVer>] - YYYY-MM-DD` 和实际发行日期开头，分类使用 `###`。源文件保留版本标题，Core API 从中提取版本与日期。GitHub Release 生成器将其替换为 `## English`、`## 简体中文`，版本由 Release 顶部标题展示。

执行 `go run ./scripts/release-notes --version 0.1.0`，按英文、简体中文顺序生成正文并去掉 HTML 注释。`scripts/dist` 将同一正文写到 `release-notes.md`。创建 GitHub Release 时用 `--notes-file` 读取该文件。更正已发布的说明时，先合入源文案，再用同一份内容重新生成并同步 GitHub Release 正文。

[English](README.md) · [v0.3.2](v0.3.2/zh-CN.md) · [v0.1.0](v0.1.0/zh-CN.md) · [发行维护](../docs/releasing.md)
