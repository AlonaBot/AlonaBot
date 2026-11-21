# Integration tests

[简体中文](README.zh-CN.md)

`agent/` contains collaboration assertions for the official Agent and a real Core. `e2e/` contains CLI, recovery and cross-component end-to-end tests. Both use a separate `go.mod`; `components.lock.json` pins component versions and Core comes from the current checkout. The Core module keeps its product dependencies independent.

Run from the repository root:

```fish
uv run --no-project python scripts/integration.py
uv run --no-project python scripts/integration.py --run '^TestAgentServices$' --race
```

`app/agent_service_test.go.tmpl` verifies Core's private assembly methods and the real Agent. The runner uses a temporary modfile and Go overlay to place it in the `internal/app` test package. Business assertions are retained and official Agent resources are scanned from the pinned project. Temporary dependency files are created and removed for that run; no test files need to be generated in the source tree.

Release first-run checks consume pure Core archives through `ALONABOT_RELEASE_DIR` and internal bundles through `ALONABOT_BUNDLE_DIR`. The release workflow sets both to verify an empty component list or the default WebUI/Agent state, login, persistence across restarts and normal shutdown. Ordinary tests do not rebuild release packages.

Platform and recovery tools reuse `e2e/` scenarios; see [scripts/platform](../scripts/platform/README.md) and [scripts/scenarios](../scripts/scenarios/README.md). Tests here use separate temporary data and local simulated services.
