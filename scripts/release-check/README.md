# Check candidate release assets and the install/update lifecycle

[简体中文](README.zh-CN.md)

`release-check` reads a candidate directory produced by `scripts/dist`, preserving extracted files and `result.json` in a new directory. It checks Core-only archives by default; `--bundle` checks internal bundles and signed components. All five targets are required by default; use `--platforms` to specify a candidate subset explicitly. Platform binaries are inspected statically.

Build the checker in the repository, then run it from any working directory. Checks require no source tree, Go toolchain or network:

```sh
go build -o /tmp/alonabot-release-check ./scripts/release-check
/tmp/alonabot-release-check --candidate /tmp/alonabot-candidate \
  --version 0.1.0 --public-key /tmp/trusted-release.pub \
  --out /tmp/alonabot-release-evidence
```

`--public-key` is a separately stored public key trusted by the caller. The official release key is [`.github/release.pub`](../../.github/release.pub).

| Option | Meaning |
| --- | --- |
| `--candidate` | Required: `dist` output directory containing release archives, `SHA256SUMS` and `install.sh`; bundle mode also requires `packages/`. |
| `--bundle` | Also check signed Agent/WebUI components and their bundled copies. |
| `--version` | Required: outer release archive version and Core build version. Each component version comes from its signed manifest. |
| `--public-key` | Required: standard base64-encoded Ed25519 public key file. Keys beside assets must match this input. |
| `--out` | Required: a new evidence directory outside the candidate directory whose parent already exists. Existing directories are preserved. |
| `--platforms` | Comma-separated targets; defaults to `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`. Must match the release archive target set. Components keep their complete signed manifests, which must include these targets. |
| `--asset-base` | Expected asset URL prefix; defaults to each component repository's `releases/download/v<component version>`. Supports custom download prefixes for historical candidates. |
| `--install-candidate` | Initial component package directory for the optional runtime chain: package the existing configdb example using a separate third-party test publisher. |
| `--update-candidate` | Package directory with the same ID, source and publisher, and a strictly higher component SemVer. |
| `--wrong-source-candidate` | Valid signed package directory with the update version, same ID/publisher, and a different source. |
| `--wrong-publisher-candidate` | Valid signed package directory with the update version, same ID/source, and a different publisher. |
| `--runtime-timeout` | Required for runtime checks, for example `2m`: deadline for the entire chain including candidate shutdown. |
| `--help` | Print actual options and exit with 0. |

Bundle mode runs steps 1–4 first; both modes run step 5:

1. Verify signatures over the exact `release.json` bytes of bundled WebUI and Agent, checking trusted publisher, component ID and bound source. Sources are `github.com/AlonaBot/alonabot-webui` and `github.com/AlonaBot/alonabot-agent` respectively.
2. Check each target package's full URL, archive length and SHA256. After safe extraction, read the entry point's actual Go build-info: packaged `build-info` must match byte for byte, the platform must match, and CGO must be disabled. Package directories contain only target packages and three release metadata files.
3. Check manifest ID, version, source, role, platform, entry point and format version against stable `pack` output; check the shared author-manifest identity embedded in the binary. The binary embeds the author manifest; the packaged manifest adds build fields. These are static checks of the official `dist/pack` format. Candidate CLI install, enable and handshake acceptance cover full YAML/Schema semantics and runtime resource registration.
4. Check any `config_schema.json`, `adapter_config_schema.json` and direct `i18n/*.json` files for valid JSON and exact embedded resource bytes. List the resources actually checked in JSON. Schemas use the text normalized by `pack`.
5. Check every release archive in `SHA256SUMS`: actual Core target, CGO settings, linked public key and version; check source and version in `source.json`. Core-only archives contain only the executable, `LICENSE` and `source.json`, with no components in their source record. In bundle mode, packages and `release.json/.sig/.pub` must match external files byte for byte. Both modes require the exact expected archive inventory.

Extraction accepts only regular files and directories, rejecting traversal, links, duplicate entries, case collisions and file/directory conflicts. Linux/macOS entry points must be executable. `install.sh` must be a nonempty regular file; release candidate integration checks cover its behavior, pinned-version compose and images.

Success returns 0; failure returns 1. stdout always contains result JSON (except for `--help`), with the full error in `error` and completed/failed checks in `checks`. Once the evidence directory exists, the same JSON is saved as `result.json`, preserving failed extraction state. For argument errors or evidence-directory creation failures, save the result from stdout. stderr contains usage or failure logs.

With only static options, `scope` is `asset-integrity` and `runtime_checks` is `not_run`. Static results record each selected platform.

To check installation and updates on the current host, also supply the four component directories and runtime deadline above. Each directory contains the platform tar.gz, `release.json`, `release.json.sig` and `release.pub` produced by official `alonabot pack`. Use the existing `examples/plugins/configdb` example with explicit `role: plugin`. Initial/update copies change only version/source in the author manifest; the wrong-source copy changes only source, and the wrong-publisher copy changes only the test signing key. Private keys remain in the build environment; the checker reads release files only.

```sh
/tmp/alonabot-release-check --candidate /tmp/alonabot-candidate \
  --version 0.1.0 --public-key /tmp/trusted-release.pub \
  --out /tmp/alonabot-lifecycle-evidence \
  --install-candidate /tmp/configdb-0.1.0 \
  --update-candidate /tmp/configdb-0.1.1 \
  --wrong-source-candidate /tmp/configdb-wrong-source \
  --wrong-publisher-candidate /tmp/configdb-wrong-publisher \
  --runtime-timeout 2m
```

`--platforms` must include the current host. After all static checks, the checker extracts the host archive again into a system temporary directory and starts it without a TTY using an independent empty cwd and data_dir, without invoking source or build tools. It obtains the first token through `admin-token show`: on Unix it checks mode 0600; on Windows it checks that the actual DACL inherited from the protected data_dir grants access only to the running account and allows reading. Tokens remain only in memory and the candidate's own private file. The management CLI uses a Bearer token file; local forwarding aligns the random listener's Host/Origin with the candidate's actual `base_url`.

Runtime checks verify the HTTP Core version and preinstalled inventory: zero components for Core-only, or the two official packages' versions, sources, publishers and default enabled states for bundles. They then install/enable third-party configdb, write `configdb_note` through the public data-management API, and read real SQLite rows and `configdb_startup` written by plugin OnStart. The management CLI must reject updates with a bad signature, valid signature but different source/publisher, or mismatched pin. After every rejection, identity, version, Active and database content are checked. Removing the pin allows a real higher-version update: Active must change and content must survive. Ordinary uninstall preserves binding and generation; reinstall from a different source/publisher is rejected. Reinstalling from the same source and enabling must restore the retained Active and original content.

Success sets `scope=asset-integrity-and-lifecycle` and `runtime_checks=pass`. `runtime` records platform, retained system temporary workspace, Core PID/exit code and component state at each stage. `out/runtime/` retains Core logs, raw management CLI stdout/stderr/arguments/exit codes and raw data API responses, including original expected-rejection errors. Failure returns the complete error and completed stages. The candidate receives an interrupt and is awaited; exceeding the caller's deadline terminates it and records failure. Temporary assets and data_dir remain for inspection. Overall success also requires clean Core exit and logs without tokens. The release matrix records actual platform runs.
