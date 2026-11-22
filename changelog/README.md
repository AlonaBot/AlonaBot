# Changelog

Version directories contain `en.md` and `zh-CN.md`. Add user-visible changes to `unreleased/`; prepare the version directory before building a release. Keep entries under Added, Changed, Deprecated, Removed, Fixed or Security, omitting empty sections. Link source PRs beside the relevant entry.

Each `v<SemVer>/{en,zh-CN}.md` file contains one Keep a Changelog version section headed by `## [<SemVer>] - YYYY-MM-DD`, using the actual release date, with `###` subsections. Keep these source headings: the Core API extracts versions and dates from them. The GitHub Release renderer replaces them with `## English` and `## 简体中文`; the Release title supplies the version.

`go run ./scripts/release-notes --version 0.1.0` renders English followed by Simplified Chinese and strips HTML comments. `scripts/dist` writes the same text to `release-notes.md`. Release creation uses that generated file through `--notes-file`; generated GitHub commit summaries are not the release body.

[简体中文](README.zh-CN.md) · [v0.1.0](v0.1.0/en.md)
