# alonabot init 模板

`alonabot init` 按组件角色复制本目录下的文件，生成新工程的骨架；`plugin.yaml` 与 `go.mod` 由 CLI 生成，不在这里。

- `common/`：所有角色共用的文件。
- `plugin/`：plugin 与 agent 角色。
- `adapter/`、`webui/`：对应角色。

文件名规则：

- 以 `_` 开头的文件名，生成时把 `_` 换成 `.`，例如 `_gitignore` 生成 `.gitignore`；这样模板本身不会影响本仓库的 Git 行为。
- 以 `.tmpl` 结尾的文件按 `text/template` 渲染，并去掉 `.tmpl` 后缀；`.go` 文件必须带 `.tmpl` 后缀，避免被当成本仓库的源码编译。可用字段为 `.ID` 与 `.AdapterID`。
- 其他文件原样复制。
