# 更新日志

每个版本目录包含 `en.md` 与 `zh-CN.md`。日常变更写入 `unreleased/`；发行准备时整理到对应版本目录。条目说明用户可见的变化，按 Added、Changed、Deprecated、Removed、Fixed、Security 分类，省略空分类，相关 PR 链接放在条目旁。

每份 `v<SemVer>/{en,zh-CN}.md` 采用 Keep a Changelog 单版本格式，以 `## [<SemVer>] - YYYY-MM-DD` 和实际发行日期开头，分类使用 `###`。生成器按语言原样拼接正式文件；Core API 从文件内标题提取版本与日期。

执行 `go run ./scripts/release-notes --version 0.1.0`，按英文、简体中文顺序生成正文并去掉 HTML 注释。`scripts/dist` 将同一正文写到 `release-notes.md`。创建 GitHub Release 时用 `--notes-file` 读取该文件。

[English](README.md) · [v0.1.0 发布说明](v0.1.0/zh-CN.md) · [发行维护](../docs/releasing.zh-CN.md)
