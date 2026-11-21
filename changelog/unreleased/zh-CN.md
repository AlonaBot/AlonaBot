# 未发布

## Added

- 提供组件与 Core 更新日志的管理 API。

## Changed

- 提供中英文 README、语言互切、各语言文档首页与共享 Views 徽章，品牌图片移至 `docs/assets`。
- plugin/agent 初始工程直接回复 `Hello AlonaBot`，按需自行添加 i18n 目录；各角色生成的 Go 文件和 `plugin.yaml` 均无注释。

## Fixed

- Core 日志 API 返回正式版本内容，按 `Accept-Language` 选择英文或简体中文，缺省或未匹配时回退英文，并排除未发布变更。
