// Package journal 编排未提交候选的准备、中止与旧 Session 恢复。
//
// Prepare 先完成静态检查，再持久化 started，停止旧写入者后依次登记数据库、
// 代次与 PREPARED Session。Result 持有候选，调用方须继续提交或调用 Abort。
// Abort 确认两棵进程树退出后才清理引用并恢复旧版；失败保留 Journal 供恢复。
// 同包操作由编排器串行化，config 与 supervisor 仍各自取得 Store 的插件锁，
// 避免持同一锁等待自身。规则见 CONTRACTS §6.3、§12.2、§12.3。
package journal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/candidate"
	"github.com/AlonaBot/AlonaBot/internal/core/config"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/secret"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/supervisor"
	"github.com/AlonaBot/AlonaBot/internal/oslock"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Kind 指定候选的配置与数据库来源。
type Kind string

const (
	// KindFirstInstall 准备第一个代次，由 SDK 创建首库。
	KindFirstInstall Kind = "first_install"
	// KindReload 导入编辑快照，并复制当前数据库。
	KindReload Kind = "reload"
	// KindUpdate 使用预检后的新包及有效 Pending 或 Active 快照。
	KindUpdate Kind = "update"
)

const (
	phaseStarted        = "started"
	phaseStoppedOld     = "stopped_old"
	phaseCandidateReady = "candidate_ready"
	phasePrepared       = "prepared"
	phaseFailed         = "failed"
	phaseCleaned        = "cleaned"
	phaseCommitted      = "committed"
	emptyConfig         = "{}"
	packageDigestPrefix = 12 // 与 packages.Candidate.MoveTo 的目录命名一致。
)

// Options 绑定同一 Core 的状态、配置和运行循环入口。
// Stop/Resume 直接在操作所有者内执行，回调不再排入 Core 的操作队列。
type Options struct {
	Store    *state.Store
	Secrets  *secret.Store
	Config   *config.Service
	Settings *settings.File
	DataDir  string
	State    func(id string) string
	Spec     func(ctx context.Context, in state.Install, packageDir string, boot *pb.Bootstrap) (supervisor.Spec, error)
	Stop     func(ctx context.Context, id string, reason pb.StopReason) error
	Resume   func(ctx context.Context, id string) error
}

// Orchestrator 持有候选准备所需入口与逐包操作互斥。
type Orchestrator struct {
	opt   Options
	locks sync.Map // 值为 *sync.Mutex；不与 Store.LockPlugin 嵌套。
}

// Result 交接同一候选代次和 PREPARED Session，内部保存清理与恢复所有权。
type Result struct {
	OpID, Generation string
	Prepared         *candidate.Prepared // NONE 后端为 nil。
	Session          *supervisor.Prepared
	record           state.Journal
	pluginDir        string
	stopAttempted    bool
	oldStopped       bool
	resumed          bool
	generationOwned  bool
	databaseOwned    bool
	cause            error // Abort 重入保留首次失败。
}

// New 返回绑定 opt 的候选编排器。
func New(opt Options) *Orchestrator { return &Orchestrator{opt: opt} }

func (o *Orchestrator) lock(id string) func() {
	value, _ := o.locks.LoadOrStore(id, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func precondition(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: message}
}

// Prepare 完成静态检查、停止旧写入者和候选准备，停在 PREPARED。
// 写 Journal 之后发生的错误连同 Result 返回，由调用方 Abort。
func (o *Orchestrator) Prepare(ctx context.Context, pluginID string, kind Kind, staged *packages.Candidate) (*Result, error) {
	unlock := o.lock(pluginID)
	defer unlock()
	// 1. 静态检查全部在 Journal 与停机之前，旧配置交给 SDK 迁移。
	if (kind != KindFirstInstall && kind != KindReload && kind != KindUpdate) || (kind == KindUpdate) != (staged != nil) {
		return nil, precondition("candidate kind and staged package do not match")
	}
	in, err := o.opt.Store.GetInstall(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	active, err := o.opt.Store.Active(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if in.Uninstalled || in.Role == state.RoleWebUI || in.Source == "dev" || strings.HasPrefix(active, "dev:") {
		return nil, precondition("component does not support candidate preparation")
	}
	if kind != KindFirstInstall && active == "" {
		return nil, precondition("candidate preparation requires an active generation")
	}
	if kind == KindFirstInstall && active != "" {
		return nil, precondition("first install already has an active generation")
	}
	records, err := o.opt.Store.Journals(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		if record.Phase != phaseCleaned && record.Phase != phaseCommitted {
			return nil, precondition("component has an unfinished candidate operation: " + record.OpID)
		}
	}
	wasEnabled, err := o.opt.Store.Enabled(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	packageDir, err := o.installedPackage(in)
	if err != nil {
		return nil, err
	}
	oldManifest, err := readManifest(packageDir)
	if err != nil {
		return nil, err
	}
	target := in
	m := oldManifest
	if staged != nil {
		packageDir = staged.Dir
		m = staged.Manifest
		target.Entrypoint, target.Version, target.PublisherKeyID = m.Entrypoint, m.Version, staged.Release.PublisherKeyID
		target.BuildDigest, err = hex.DecodeString(staged.Asset.SHA256)
		if err != nil {
			return nil, err
		}
	}
	if kind != KindFirstInstall && (oldManifest.PrepareEffects == manifest.EffectsExternal || m.PrepareEffects == manifest.EffectsExternal) {
		return nil, precondition("NEEDS_MANUAL_MIGRATION")
	}
	if in.Role == state.RoleAdapter {
		if err := o.trialInstances(in, packageDir); err != nil {
			return nil, err
		}
	}
	configuration, err := o.configuration(ctx, pluginID, active, kind)
	if err != nil {
		return nil, err
	}
	boot := &pb.Bootstrap{GenerationId: "gen-" + rand.Text(), CandidateConfigJson: string(configuration), TargetConfigVersion: *m.ConfigVersion,
		DatabaseDescriptor: &pb.DatabaseDescriptor{ActivationReason: activationReason(kind), ManualMigrationApproved: false}}
	if active != "" {
		version, err := o.opt.Store.GenerationConfigVersion(ctx, active)
		if err != nil {
			return nil, err
		}
		boot.PreviousConfigVersion = &version
	}
	spec, err := o.opt.Spec(ctx, target, packageDir, boot)
	if err != nil {
		return nil, err
	}
	spec.NewCandidate, spec.ExpectedOld = true, active
	spec.Secrets = o.opt.Secrets
	pluginDir, err := filepath.Abs(spec.PluginDir)
	if err != nil {
		return nil, err
	}
	// 2. 记录操作身份，再取得新包的清理所有权；安装记录始终保持旧程序身份。
	r := &Result{OpID: "op-" + rand.Text(), Generation: boot.GenerationId, pluginDir: pluginDir}
	r.record = state.Journal{OpID: r.OpID, PluginID: pluginID, Kind: string(kind), ExpectedOld: active, NewGeneration: r.Generation, Phase: phaseStarted,
		PackageDir: packageDir, Entrypoint: target.Entrypoint, Version: target.Version, PublisherKeyID: target.PublisherKeyID, BuildDigest: target.BuildDigest,
		WasRunning: o.opt.State(pluginID) != "", WasEnabled: wasEnabled}
	if err := o.opt.Store.CreateJournal(ctx, r.record); err != nil {
		return nil, err
	}
	if staged != nil {
		relative, err := filepath.Rel(packageDir, spec.Executable)
		if err != nil {
			return r, err
		}
		final, err := staged.MoveTo(filepath.Join(o.opt.DataDir, "packages"))
		if err != nil {
			return r, err
		}
		r.record.PackageDir, r.record.OwnsPackage = final, true
		spec.Executable = filepath.Join(final, relative)
		// 即使请求已取消，取得的程序引用仍须登记，供 Abort 或崩溃恢复处理。
		syncErr := persistPackage(final)
		if err := errors.Join(syncErr, o.save(context.WithoutCancel(ctx), r, phaseStarted)); err != nil {
			return r, err
		}
	}
	// 3. 原停止链负责冻结、排空与撤销 Session；返回后再核对跨代次系统锁。
	r.stopAttempted = true
	if err := o.opt.Stop(ctx, pluginID, stopReason(kind)); err != nil {
		return r, err
	}
	r.oldStopped = true
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		return r, err
	}
	if err := checkRuntime(pluginDir); err != nil {
		return r, err
	}
	if err := o.save(ctx, r, phaseStoppedOld); err != nil {
		return r, err
	}
	// 4. 数据库独立准备后登记代次。初始快照使用空对象，避免明文落入状态库。
	database := spec.Bootstrap.GetDatabaseDescriptor()
	if database.GetBackend() == pb.DatabaseBackend_DATABASE_BACKEND_SQLITE {
		r.Prepared, err = candidate.PrepareDatabase(ctx, candidate.Request{PluginID: pluginID, PluginDir: pluginDir, Reason: activationReason(kind), FromGeneration: active, NewGeneration: r.Generation})
		if err != nil {
			// PrepareDatabase 清理失败时，仍登记本操作独占代次目录中的残留资源。
			path := filepath.Join(pluginDir, "generations", r.Generation, "db.sqlite")
			if _, statErr := os.Stat(filepath.Dir(path)); statErr == nil {
				r.record.DatabasePath = path
				r.databaseOwned = true
				err = errors.Join(err, o.save(context.WithoutCancel(ctx), r, phaseStoppedOld))
			} else if !errors.Is(statErr, os.ErrNotExist) {
				err = errors.Join(err, statErr)
			}
			return r, err
		}
		r.databaseOwned = true
		r.record.DatabasePath = r.Prepared.DatabasePath
		database.SqlitePath = r.Prepared.DatabasePath
		if err := o.save(context.WithoutCancel(ctx), r, phaseStoppedOld); err != nil {
			return r, err
		}
	}
	if err := o.opt.Store.AddGeneration(ctx, pluginID, r.Generation, emptyConfig, boot.TargetConfigVersion); err != nil {
		return r, err
	}
	r.generationOwned = true
	// Pending、旧代次与其他候选的快照持续保护已有 secret 引用。
	if err := o.save(ctx, r, phaseCandidateReady); err != nil {
		return r, err
	}
	// 5. supervisor 只运行到 PREPARED；候选不进入普通异常恢复或业务装配。
	r.Session, err = supervisor.Prepare(ctx, o.opt.Store, spec)
	if err != nil {
		return r, err
	}
	r.record.SessionID = append([]byte(nil), r.Session.Registration.SessionID...)
	if r.Prepared != nil {
		if err := persistDatabase(r.Prepared.DatabasePath, kind == KindFirstInstall); err != nil {
			return r, err
		}
	}
	if err := o.save(ctx, r, phasePrepared); err != nil {
		return r, err
	}
	return r, nil
}

func (o *Orchestrator) save(ctx context.Context, r *Result, phase string) error {
	next := r.record
	next.Phase = phase
	if err := o.opt.Store.UpdateJournal(ctx, next); err != nil {
		return err
	}
	r.record = next
	return nil
}

func activationReason(kind Kind) pb.ActivationReason {
	switch kind {
	case KindFirstInstall:
		return pb.ActivationReason_ACTIVATION_REASON_FIRST_INSTALL
	case KindReload:
		return pb.ActivationReason_ACTIVATION_REASON_RELOAD
	default:
		return pb.ActivationReason_ACTIVATION_REASON_UPDATE
	}
}

func stopReason(kind Kind) pb.StopReason {
	if kind == KindUpdate {
		return pb.StopReason_STOP_REASON_UPDATE
	}
	return pb.StopReason_STOP_REASON_RELOAD
}

func (o *Orchestrator) installedPackage(in state.Install) (string, error) {
	dir, err := filepath.Abs(o.opt.DataDir)
	if err != nil {
		return "", err
	}
	digest := hex.EncodeToString(in.BuildDigest)
	if len(digest) < packageDigestPrefix {
		return "", precondition("installed build digest is invalid")
	}
	return filepath.Join(dir, "packages", in.PluginID, in.Version+"-"+digest[:packageDigestPrefix]), nil
}

func readManifest(dir string) (*manifest.Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "plugin.yaml"))
	if err != nil {
		return nil, err
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return nil, err
	}
	return parsed.Manifest, nil
}

func (o *Orchestrator) trialInstances(in state.Install, dir string) error {
	raw, err := os.ReadFile(filepath.Join(dir, "adapter_config_schema.json"))
	if err != nil {
		return err
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) != 0 {
		return precondition(fmt.Sprintf("adapter instance schema: %v", problems))
	}
	group := o.opt.Settings.Settings().Adapters[in.AdapterID]
	if group.Package != "" && group.Package != in.PluginID {
		return nil
	}
	var diagnostics []string
	for _, problem := range o.opt.Settings.TrialInstances(in.AdapterID, schema) {
		diagnostics = append(diagnostics, fmt.Sprintf("instance %s at %s: %s", problem.InstanceID, problem.Path, problem.Reason))
	}
	if len(diagnostics) != 0 {
		return precondition(strings.Join(diagnostics, "; "))
	}
	return nil
}

func (o *Orchestrator) configuration(ctx context.Context, id, active string, kind Kind) ([]byte, error) {
	if kind == KindReload {
		pending, err := o.opt.Config.ImportSnapshot(ctx, id)
		return pending.ConfigJSON, err
	}
	if kind == KindUpdate {
		pending, found, err := o.opt.Config.Pending(ctx, id)
		if err != nil {
			return nil, err
		}
		if found {
			return pending.ConfigJSON, nil
		}
		snapshot, err := o.opt.Store.GenerationConfig(ctx, active)
		if err != nil {
			return nil, err
		}
		return config.Canonical([]byte(snapshot))
	}
	raw, err := os.ReadFile(filepath.Join(o.opt.DataDir, "plugins", id, "config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return []byte(emptyConfig), nil
	}
	if err != nil {
		return nil, err
	}
	return config.Canonical(raw)
}

func checkRuntime(dir string) error {
	lock, err := oslock.TryLock(filepath.Join(dir, wire.RuntimeLockFile))
	if errors.Is(err, oslock.ErrLocked) {
		return precondition("plugin process still holds runtime.lock")
	}
	if err != nil {
		return err
	}
	return lock.Unlock()
}
