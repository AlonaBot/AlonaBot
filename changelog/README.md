# Changelog

Version directories contain `en.md` and `zh-CN.md`. Add user-visible changes to `unreleased/`; prepare the version directory before building a release. Keep entries under Added, Changed, Deprecated, Removed, Fixed or Security, omitting empty sections. Link source PRs beside the relevant entry.

`go run ./scripts/release-notes --version 0.1.0` renders English followed by Simplified Chinese and strips HTML comments. `scripts/dist` writes the same text to `release-notes.md`. Release creation uses that generated file through `--notes-file`; generated GitHub commit summaries are not the release body.

[简体中文](README.zh-CN.md) · [v0.1.0 candidate](v0.1.0/en.md)
