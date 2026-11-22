package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/release"
)

type runtimeResult struct {
	Platform  string                  `json:"platform"`
	Workspace string                  `json:"workspace"` // 保留源码外 cwd、data_dir 与实际使用资产。
	PID       int                     `json:"pid"`
	ExitCode  *int                    `json:"exit_code,omitempty"`
	Packages  map[string]packageState `json:"packages"`
}

// packageState 对应公开 packages API；保留本联验涉及的身份与代次字段。
type packageState struct {
	ID, Role, AdapterID, Version, Source, PublisherKeyID, Pin string
	Enabled, Uninstalled, Signed                              bool
	Active, State                                             string
}

type runtimeCandidate struct {
	publication        *manifest.Manifest
	directory, archive string
}

func verifyLifecycle(opt options, r *report) (err error) {
	target := runtime.GOOS + "/" + runtime.GOARCH
	if !slices.Contains(r.Platforms, target) {
		return fmt.Errorf("runtime checks require host platform %s in --platforms", target)
	}
	workspace, err := os.MkdirTemp("", "alonabot-release-run-")
	if err != nil {
		return err
	}
	r.Runtime = &runtimeResult{Platform: target, Workspace: workspace, Packages: map[string]packageState{}}
	evidence := filepath.Join(opt.out, "runtime")
	if err := os.Mkdir(evidence, directoryMode); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), opt.runtimeTimeout)
	defer cancel()
	var initial, updated, wrongSource *runtimeCandidate
	if err := r.perform("runtime/inputs", func() error {
		for _, input := range []struct {
			name, dir   string
			destination **runtimeCandidate
		}{
			{"initial", opt.installCandidate, &initial}, {"updated", opt.updateCandidate, &updated},
			{"wrong-source", opt.wrongSource, &wrongSource},
		} {
			candidate, err := prepareRuntimeCandidate(input.dir, filepath.Join(workspace, input.name))
			if err != nil {
				return fmt.Errorf("%s: %w", input.name, err)
			}
			*input.destination = candidate
		}
		old, next := initial.publication, updated.publication
		if old.ID != "alonabot-plugin-configdb" || next.ID != old.ID || next.Source != old.Source || compareSemVer(next.Version, old.Version) <= 0 {
			return errors.New("initial/update must be configdb example packages with the same identity/source and a strictly increasing component SemVer")
		}
		other := wrongSource.publication
		if other.ID != old.ID || other.Source == old.Source || other.Version != next.Version {
			return errors.New("wrong-source release must match updated ID/version and differ in source")
		}
		return nil
	}); err != nil {
		return err
	}
	// 1. 从同一发行归档再次解包到系统临时目录，运行 cwd/data_dir 均为新目录。
	rootName := fmt.Sprintf("alonabot-%s-%s-%s", opt.version, runtime.GOOS, runtime.GOARCH)
	extension, executable := ".tar.gz", "alonabot"
	if runtime.GOOS == "windows" {
		extension, executable = ".zip", "alonabot.exe"
	}
	distribution := filepath.Join(workspace, "distribution")
	if _, err := extractArchive(filepath.Join(opt.candidate, rootName+extension), distribution); err != nil {
		return err
	}
	cwd, data := filepath.Join(workspace, "cwd"), filepath.Join(workspace, "data")
	if err := os.Mkdir(cwd, directoryMode); err != nil {
		return err
	}
	var host *runtimeHost
	defer func() {
		if host == nil {
			return
		}
		err = errors.Join(err, r.perform("runtime/stop", host.stop))
		if host.command.ProcessState != nil {
			code := host.command.ProcessState.ExitCode()
			r.Runtime.ExitCode = &code
		}
		if host.token != "" {
			err = errors.Join(err, r.perform("runtime/token-log-secrecy", func() error {
				for _, name := range []string{"core.stdout.log", "core.stderr.log"} {
					raw, err := os.ReadFile(filepath.Join(evidence, name))
					if err != nil {
						return err
					}
					if bytes.Contains(raw, []byte(host.token)) {
						return fmt.Errorf("admin token appeared in %s", name)
					}
				}
				return nil
			}))
		}
	}()
	if err := r.perform("runtime/start", func() error {
		var startErr error
		host, startErr = startRuntime(ctx, filepath.Join(distribution, rootName, executable), cwd, data, evidence, opt.bundle)
		if host != nil {
			r.Runtime.PID = host.command.Process.Pid
		}
		return startErr
	}); err != nil {
		return err
	}
	if err := r.perform("runtime/first-token", host.authenticate); err != nil {
		return err
	}
	closeManagement, err := host.serveManagement()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, closeManagement()) }()
	if err := r.perform("runtime/preinstalled", func() error {
		raw, err := host.api("core-version", http.MethodGet, "/api/v1/core", "")
		if err != nil {
			return err
		}
		var core struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(raw, &core); err != nil {
			return err
		}
		if core.Version != opt.version {
			return fmt.Errorf("HTTP Core version %q; want %q", core.Version, opt.version)
		}
		packages, err := host.list("preinstalled")
		if err != nil {
			return err
		}
		if len(packages) != len(r.Components) {
			return fmt.Errorf("preinstalled count = %d; want %d", len(packages), len(r.Components))
		}
		for _, expected := range r.Components {
			item, ok := packages[expected.ID]
			enabled := expected.ID == "alonabot-webui-admin"
			if !ok || item.Version != expected.Version || item.Source != expected.Source || item.Uninstalled || item.Enabled != enabled {
				return fmt.Errorf("unexpected preinstalled state: %+v", item)
			}
			if enabled && (item.State != "Activated" || item.Active == "") || !enabled && (item.State != "Installed" || item.Active != "") {
				return fmt.Errorf("unexpected preinstalled activation: %+v", item)
			}
			r.Runtime.Packages["preinstalled/"+item.ID] = item
		}
		return nil
	}); err != nil {
		return err
	}
	id := initial.publication.ID
	var active packageState
	if err := r.perform("runtime/install-and-seed", func() error {
		if _, err := host.cli("install", initial.args("install", id)...); err != nil {
			return err
		}
		if _, err := host.cli("enable", "enable", id, "--yes"); err != nil {
			return err
		}
		var err error
		active, err = host.snapshot("installed", id, r)
		if err != nil {
			return err
		}
		if err := checkActive(active, initial.publication); err != nil {
			return err
		}
		if _, err := host.api("seed-database", http.MethodPost, "/api/v1/plugins/"+id+"/data/tables/configdb_note/rows", `{"values":{"key":"release-check","value":"preserved-through-update-and-reinstall"}}`); err != nil {
			return err
		}
		return host.checkDatabase("installed", id)
	}); err != nil {
		return err
	}
	// 2. 每次拒绝均核对完整身份/Active 快照和真实 SQLite 内容。
	badArchive := filepath.Join(workspace, "invalid-package.tar.gz")
	if err := os.WriteFile(badArchive, []byte("invalid gzip archive"), fileMode); err != nil {
		return err
	}
	for _, rejection := range []struct {
		name, message string
		args          []string
	}{
		{"bad-archive", "archive is not valid gzip data", []string{"update", id, "--archive", badArchive, "--yes"}},
		{"wrong-source", "candidate source differs from installed binding", wrongSource.args("update", id)},
	} {
		if err := r.perform("runtime/reject-"+rejection.name, func() error {
			return host.rejectUnchanged(rejection.name, rejection.message, rejection.args, active, r, true)
		}); err != nil {
			return err
		}
	}
	if err := r.perform("runtime/pin", func() error {
		if _, err := host.cli("pin", "pin", id, active.Version, "--yes"); err != nil {
			return err
		}
		active.Pin = active.Version
		return host.rejectUnchanged("pinned-update", "candidate version does not match pin", updated.args("update", id), active, r, true)
	}); err != nil {
		return err
	}
	if err := r.perform("runtime/update", func() error {
		if _, err := host.cli("unpin", "pin", id, "--yes"); err != nil {
			return err
		}
		if _, err := host.cli("update", updated.args("update", id)...); err != nil {
			return err
		}
		next, err := host.snapshot("updated", id, r)
		if err != nil {
			return err
		}
		if err := checkActive(next, updated.publication); err != nil {
			return err
		}
		if next.Active == active.Active || next.Pin != "" {
			return fmt.Errorf("update did not change Active or clear pin: %+v", next)
		}
		active = next
		return host.checkDatabase("updated", id)
	}); err != nil {
		return err
	}
	// 3. 普通卸载保留代次、数据与绑定，错误重装拒绝后再同源重装。
	if err := r.perform("runtime/uninstall", func() error {
		if _, err := host.cli("uninstall", "uninstall", id, "--yes"); err != nil {
			return err
		}
		removed, err := host.snapshot("uninstalled", id, r)
		if err != nil {
			return err
		}
		want := active
		want.Enabled, want.Uninstalled, want.State = false, true, "Stopped"
		if removed != want {
			return fmt.Errorf("uninstall changed retained identity/generation: got %+v; want %+v", removed, want)
		}
		active = removed
		return nil
	}); err != nil {
		return err
	}
	for _, rejection := range []struct {
		name, message string
		candidate     *runtimeCandidate
	}{
		{"reinstall-wrong-source", "reinstallation source or identity differs from retained binding", wrongSource},
	} {
		if err := r.perform("runtime/"+rejection.name, func() error {
			return host.rejectUnchanged(rejection.name, rejection.message, rejection.candidate.args("install", id), active, r, false)
		}); err != nil {
			return err
		}
	}
	return r.perform("runtime/reinstall", func() error {
		if _, err := host.cli("reinstall", updated.args("install", id)...); err != nil {
			return err
		}
		if _, err := host.cli("reenable", "enable", id, "--yes"); err != nil {
			return err
		}
		reinstalled, err := host.snapshot("reinstalled", id, r)
		if err != nil {
			return err
		}
		if err := checkActive(reinstalled, updated.publication); err != nil {
			return err
		}
		if reinstalled.Active != active.Active {
			return fmt.Errorf("reinstall changed retained Active: got %q; want %q", reinstalled.Active, active.Active)
		}
		return host.checkDatabase("reinstalled", id)
	})
}

func prepareRuntimeCandidate(dir, destination string) (*runtimeCandidate, error) {
	if err := os.Mkdir(destination, directoryMode); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	publication := release.GitHubRelease{}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			publication.Assets = append(publication.Assets, release.Download{Name: entry.Name()})
		}
	}
	asset, err := publication.Select(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}
	raw, err := readRegular(filepath.Join(dir, asset.Name))
	if err != nil {
		return nil, err
	}
	archive := filepath.Join(destination, asset.Name)
	if err := os.WriteFile(archive, raw, fileMode); err != nil {
		return nil, err
	}
	unpacked := filepath.Join(destination, "unpacked")
	files, err := extractArchive(archive, unpacked)
	if err != nil {
		return nil, err
	}
	raw, err = readRegular(filepath.Join(unpacked, "plugin.yaml"))
	if err != nil {
		return nil, err
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return nil, err
	}
	m := parsed.Manifest
	if release.ArchiveName(m) != asset.Name {
		return nil, errors.New("runtime asset name differs from package manifest")
	}
	if _, err := checkPackage(unpacked, files, component{m.ID, "plugin", m.Source}, m.Version, runtime.GOOS, runtime.GOARCH); err != nil {
		return nil, err
	}
	return &runtimeCandidate{m, destination, archive}, nil
}

func (c *runtimeCandidate) args(action, id string) []string {
	args := []string{action}
	if action == "update" {
		args = append(args, id, "--archive", c.archive)
	} else {
		args = append(args, c.archive)
	}
	return append(args, "--yes")
}

func (h *runtimeHost) list(name string) (map[string]packageState, error) {
	raw, err := h.cli(name, "list")
	if err != nil {
		return nil, err
	}
	var packages []packageState
	if err := json.Unmarshal(raw, &packages); err != nil {
		return nil, err
	}
	result := map[string]packageState{}
	for _, item := range packages {
		if _, exists := result[item.ID]; exists {
			return nil, fmt.Errorf("duplicate package ID %q in list", item.ID)
		}
		result[item.ID] = item
	}
	return result, nil
}

func (h *runtimeHost) snapshot(name, id string, r *report) (packageState, error) {
	packages, err := h.list(name)
	if err != nil {
		return packageState{}, err
	}
	item, ok := packages[id]
	if !ok {
		return packageState{}, fmt.Errorf("package %s is missing after %s", id, name)
	}
	r.Runtime.Packages[name] = item
	return item, nil
}

func checkActive(item packageState, publication *manifest.Manifest) error {
	if item.ID != publication.ID || item.Role != "plugin" || item.Version != publication.Version || item.Source != publication.Source || !item.Enabled || item.Uninstalled || item.State != "Activated" || item.Active == "" {
		return fmt.Errorf("unexpected active package: %+v", item)
	}
	return nil
}

func (h *runtimeHost) rejectUnchanged(name, message string, args []string, before packageState, r *report, database bool) error {
	_, err := h.cli(name, args...)
	if err == nil {
		return fmt.Errorf("%s unexpectedly succeeded", name)
	}
	var evidenceErr *cliEvidenceError
	if errors.As(err, &evidenceErr) {
		return err
	}
	if !strings.Contains(err.Error(), message) {
		return fmt.Errorf("%s: expected rejection %q: %w", name, message, err)
	}
	after, err := h.snapshot(name+"-unchanged", before.ID, r)
	if err != nil {
		return err
	}
	if after != before {
		return fmt.Errorf("%s changed state: before %+v; after %+v", name, before, after)
	}
	if database {
		return h.checkDatabase(name, before.ID)
	}
	return nil
}

func (h *runtimeHost) checkDatabase(name, id string) error {
	for _, table := range []struct{ table, key, value string }{
		{"configdb_note", "release-check", "preserved-through-update-and-reinstall"},
		{"configdb_startup", "startup", "Hello"},
	} {
		raw, err := h.api(name+"-"+table.table, http.MethodGet, "/api/v1/plugins/"+id+"/data/tables/"+table.table+"/rows?where.key="+table.key, "")
		if err != nil {
			return err
		}
		var result struct {
			Total int `json:"total"`
			Rows  []struct {
				Values map[string]any `json:"values"`
			} `json:"rows"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			return err
		}
		if result.Total != 1 || len(result.Rows) != 1 || result.Rows[0].Values["key"] != table.key || result.Rows[0].Values["value"] != table.value {
			return fmt.Errorf("%s database content differs: %s", table.table, raw)
		}
	}
	return nil
}

// compareSemVer 只接收 manifest.Parse 已验证的版本，忽略 build metadata。
func compareSemVer(a, b string) int {
	parts := func(value string) ([]string, string) {
		value, _, _ = strings.Cut(value, "+")
		base, pre, _ := strings.Cut(value, "-")
		return strings.Split(base, "."), pre
	}
	aBase, aPre := parts(a)
	bBase, bPre := parts(b)
	numeric := func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b)
	}
	for i := range aBase {
		if c := numeric(aBase[i], bBase[i]); c != 0 {
			return c
		}
	}
	if aPre == bPre {
		return 0
	}
	if aPre == "" {
		return 1
	}
	if bPre == "" {
		return -1
	}
	aParts, bParts := strings.Split(aPre, "."), strings.Split(bPre, ".")
	for i := 0; i < min(len(aParts), len(bParts)); i++ {
		x, y := aParts[i], bParts[i]
		xNumeric, yNumeric := strings.Trim(x, "0123456789") == "", strings.Trim(y, "0123456789") == ""
		if xNumeric && yNumeric {
			if c := numeric(x, y); c != 0 {
				return c
			}
		} else if xNumeric != yNumeric {
			if xNumeric {
				return -1
			}
			return 1
		} else if c := strings.Compare(x, y); c != 0 {
			return c
		}
	}
	return len(aParts) - len(bParts)
}
