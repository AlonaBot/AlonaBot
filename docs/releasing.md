# Release maintenance

[简体中文](releasing.zh-CN.md)

Core releases and bundled images remain in the main repository. Public five-platform archives contain only Core; Docker Compose images contain Core, WebUI and Agent. Telegram, OneBot 11 and help release from their own repositories and are installed as needed. See [deployment](deployment.md) for installation and the [plugin guide](plugin-guide.md#9-publishing) for component authors.

## Pin source and release notes

1. Select a reviewed Core commit and release version, and confirm the corresponding checks passed.
2. Pin exact Go module versions for all five official components in `components.lock.json`. Bundle builds consume Agent and WebUI; cross-component tests also use the other entries. Component versions come from their own `plugin.yaml` and can advance independently of Core.
3. Prepare user-visible changes in `changelog/v<version>/en.md` and `zh-CN.md` with one version heading and the actual release date, following the [changelog format](../changelog/README.md). Run `go generate ./internal/core/manage` to embed both languages unchanged in the Core management API. Retain real platform results, failed resource measurements and approved deferrals. Daily entries go in `changelog/unreleased/`.
4. Pin component SDKs to a verified exact Core commit first, then pin the integration lock and Core tag after component publication. This avoids references to unpublished versions. Update the plugin template to the official SDK tag afterward.

Generated Release notes place English before Simplified Chinese and remove HTML editing comments:

```fish
go run ./scripts/release-notes --version 0.1.0 > /tmp/alonabot-release-notes.md
```

## Build Core and internal bundles

Run from the repository root:

```fish
go run ./scripts/dist --version 0.1.0 \
  --public-key .github/release.pub --out ./dist
```

Building requires Go and network access. The default builds only Core from the current checkout. Add `--bundle` explicitly for internal image bundles: it downloads public Agent and WebUI Releases pinned in `components.lock.json`, verifies signatures with the trusted key, then checks source, ID, version, platform, package length and manifest digest. Signed manifests and component packages are included byte for byte. WebUI builds its frontend during its own release. Default targets are `linux/amd64,linux/arm64,windows/amd64,darwin/amd64,darwin/arm64`; select a subset with `--platforms linux/amd64,linux/arm64`.

`--public-key` points to an independently trusted, standard base64 Ed25519 public key. Official releases use [`.github/release.pub`](../.github/release.pub). Bundling needs only the public key. Component `pack --key` accepts the maintainer's 64-byte private key; local backups use mode 0600.

Each platform produces `alonabot-<version>-<os>-<arch>.tar.gz`, or `.zip` for Windows. Default archives contain only Core, the AGPL license and `source.json` recording source versions. `--bundle` additionally includes `packages/<componentID>/`, with the target package and complete `release.json`, `release.json.sig` and `release.pub`. Source records list the actual bundled Agent and WebUI. The WebUI package contains its backend process and the same frontend build.

In bundle mode, `<output>/packages/<componentID>/` collects selected platform packages and the three metadata files. Default asset URLs point to the component repository's `releases/download/v<componentVersion>`. `--version` injects the Core version, available through `GET /api/v1/core`.

Core-only archives can be built first. Build internal bundles and Compose images after the pinned Agent and WebUI platform packages and signature metadata are public. Store acceptance uses an anonymously readable public index and a released WebUI; see [store operations](deployment.md#component-store) for interfaces and submissions.

Public Releases upload only five platform archives; GitHub provides source zip and tar.gz automatically. Download the installer from `main/deploy/install.sh`, Compose from the corresponding tag's root, and the official public key from `.github/release.pub`. The installer verifies official archives using GitHub Release API asset digests.

Build and acceptance outputs also contain these auxiliary files, retained in Actions artifacts:

- `SHA256SUMS`: integrity digests for archives in that output directory.
- `install.sh`: system-service installation and upgrade entry point.
- `docker-compose.yml`: installation entry point for a pinned bundled image.
- `release-notes.md`: official text generated from the version's bilingual changelog.

## Build container images

The release workflow writes five-platform Core archives to `dist/core` and Linux amd64/arm64 internal bundles to `dist/bundle`, checking integrity and first startup separately. Public Releases upload only the five `dist/core` archives. Native runners for the two architectures build the `published` target from their respective `dist/bundle` archives. The root Dockerfile's default target can still build dist from source:

```fish
set version 0.1.0
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION=$version \
  --output type=oci,dest=/tmp/alonabot-$version.oci.tar \
  -t ghcr.io/alonabot/alonabot:$version .
```

Use `--load` for local single-platform acceptance. The release workflow runs `scripts/release-smoke.py --compose docker-compose.yml --image <candidate-image>` with a separate project and volume, verifies login, default components, persistent credentials and shutdown, then removes that project.

Images use scratch, UID/GID 10001, a separate `/data` volume and init, with ports published only on loopback. Build an image containing the required runtimes for MCP stdio tools that need shell, Node or Python.

## Check candidates

[release-check](../scripts/release-check/README.md) checks the full Core archive inventory, source, target platform and build-info by default. `--bundle` also verifies both components' signed metadata, package contents and build-info, saving individual results in a new directory. Choose targets and lifecycle checks according to the change and reuse applicable historical evidence.

Run cross-component tests with `uv run --no-project python scripts/integration.py`. Use [platform tools](../scripts/platform/README.md) and [scenario tools](../scripts/scenarios/README.md) for platform, crash-recovery and installation/update checks. Record cross-compilation separately from actual execution. Windows runtime checks were approved for deferral; see [v0.1.0 release notes](../changelog/v0.1.0/en.md).

After changing `install.sh`, run `sh -n deploy/install.sh`. Verify real systemd/launchd installation, upgrades, verification failures and removal in an authorized, recoverable environment.

## Publish components and bundled releases

Each component's same-version Release uploads platform packages plus fixed-name `release.json`, `release.json.sig` and `release.pub`. Separate repositories prevent manifest-name collisions. Source, component ID, signing key and archive digest must agree. Verify public downloads and subsequent updates with `alonabot install --source AlonaBot/<component-repository>`.

Publish in this order:

1. Pin component SDKs to a reviewed exact Core commit; verify and merge source PRs for all five components.
2. Create `v<version>` tags in the five repositories to trigger their Release workflows. Each builds five-platform packages from pinned source, signs with the official key and reads notes from `changelog/<version>/en-US.md` and `zh-CN.md`.
3. Pin published versions in the main repository's `components.lock.json`. Verify and merge, then create the Core `v<version>` tag. The workflow builds five-platform Core archives and two Linux internal bundles, verifies archives, signatures and licenses, and checks both first-start modes before creating a Release draft containing only Core archives. Notes come from `dist/core/release-notes.md`.
4. Native amd64 and arm64 runners unpack the current internal bundle, build the Dockerfile's `published` target, and check login, default components, persistent volume credentials and shutdown with release Compose. They then push platform images and combine the version tag and `latest`.
5. For the first release, allow Public package creation under organization Settings → Packages → Package creation, then set GHCR visibility to public. After anonymous pulls and public Compose acceptance for both architectures, the workflow publishes the Release. Update README, synchronize dev and switch the plugin template to the official SDK tag.

The official public key is [`.github/release.pub`](../.github/release.pub). All five component workflows read the same publisher private key through the `ALONABOT_RELEASE_KEY` secret; local backups use mode 0600. Workflows log only the public key and signature artifacts.

On publication failure, retain the draft, complete errors, artifacts and current live version. Repair affected steps and continue. `workflow_dispatch` accepts an existing `tag` to resume an unfinished release. Ordinary releases preserve public assets and tags. On 2026-11-21, maintainers approved replacing v0.1.0's five bundled archives in place with pure Core archives, reusing the original Core binaries and retaining the tag and bundled images while regenerating digests and notes. Old archive digests became invalid; users must download again and verify Release API digests. That repack manually consumed already-verified artifacts. The old tag's workflow remains unchanged; the new workflow applies to later source tags.
