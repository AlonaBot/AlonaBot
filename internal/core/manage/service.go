// Package manage 提供组件生命周期、配置、权限、工具、会话、日志与诊断包的管理服务。
//
// [Service] 共用 Core 的状态库、设置文件与策略；[Service.Handler] 经管理认证保护，
// [Service.WebUI] 提供登录前可读取的静态页面。配置读取隐藏秘密，实例配置写回共享设置文件。
// 管理、日志与静态组件语义见 CONTRACTS §7.5、§9.3、§14、§15.1。
package manage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/adapter"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/options"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/policy"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/scheduler"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/core/tool"
)

// Options 绑定同一个 Core 的持久状态与运行登记。
type Options struct {
	Logs               *logging.Output                                               // 与 Core 共用的输出、缓冲与级别表。
	Stopping           <-chan struct{}                                               // Core 开始停机时关闭，在 HTTP Shutdown 等待之前结束日志流。
	NotifyLogLevel     func(ctx context.Context, id string, level pb.LogLevel) error // 仅通知当前 Activated Session。
	Store              *state.Store
	Settings           *settings.File
	Policy             *policy.Policy
	Router             *router.Router
	Options            *options.Service
	Tools              *tool.Catalog
	Adapters           *adapter.Registry
	State              func(id string) string                                             // 必填；返回运行状态，未运行时返回空串。
	InstanceSchema     func(ctx context.Context, id string) (*configschema.Schema, error) // 读取包内的实例声明。
	Packages           string                                                             // 不可变程序包目录，其父目录是 Core data_dir。
	Limits             packages.Limits
	ReleaseBaseURL     string                                                      // 缺省 https://github.com，测试可指向回环服务。
	HTTPClient         *http.Client                                                // nil 使用 http.DefaultClient。
	Development        map[string]*manifest.Manifest                               // 由库模式登记的只读工程身份。
	DevelopmentConfig  func(id string) (json.RawMessage, map[string][]byte, error) // 读取当前开发资源；返回值只读。
	RestartDevelopment func(context.Context, string) error                         // 配置落盘后沿 Core 运行循环重启。
	Start              func(context.Context, string) error
	Stop               func(context.Context, string, pb.StopReason) error
	StartInstance      func(context.Context, string, string, json.RawMessage) error
	StopInstance       func(context.Context, string, string) error // 与组件恢复串行，并保留本次 Core 运行的停用选择。
	ForgetInstance     func(context.Context, string, string) error // 配置创建、删除成功后清除该实例的运行期选择。
	Descriptor         func(string) *pb.AdapterDescriptor          // 返回当前运行 Session 的 Adapter 声明。
	Candidates         Candidates

	DevelopmentWebUI fs.FS                           // 本次运行的静态根目录；nil 时选择已启用的安装包。
	CoreInfo         func() (version, listen string) // 总览在请求时读取同一个 Core 的构建版本与实际监听地址。

	Scheduler             *scheduler.Scheduler
	InstanceStatus        func(ctx context.Context, packageID, instanceID string) (state, safeError string, err error) // 远端实例的运行状态；直接装配 Registry 时可省略。
	RecordInstanceFailure func(ctx context.Context, packageID, instanceID, safeError string) error                     // 可选，留存启动前的配置失败。
}

// Service 用独立的逐包锁串行管理动作，supervisor 继续持有状态库的进程锁。
type Service struct {
	logs      sync.Mutex // 串行提交日志设置和通知。
	opt       Options
	admins    sync.Mutex
	instances sync.Mutex // 实例名跨包唯一，串行化配置创建与删除。
	packages  sync.Map   // 包 ID → *sync.Mutex；锁实例在服务生命周期内保持稳定。
	catalog   sync.Mutex // 安装、同类替换、卸载与依赖图写入共用一个顺序。
}

// New 创建复用 Core 装配资源的管理服务。
func New(opt Options) *Service {
	if opt.ReleaseBaseURL == "" {
		opt.ReleaseBaseURL = "https://github.com"
	}
	if opt.HTTPClient == nil {
		opt.HTTPClient = http.DefaultClient
	}
	if opt.Candidates == nil {
		opt.Candidates = unsupportedCandidates{}
	}
	return &Service{opt: opt}
}

func (s *Service) lock(id string) func() {
	value, _ := s.packages.LoadOrStore(id, &sync.Mutex{})
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	return mutex.Unlock
}

func (s *Service) directory(in state.Install) string {
	return filepath.Join(s.opt.Packages, in.PluginID, in.Version+"-"+hex.EncodeToString(in.BuildDigest)[:12])
}

func (s *Service) manifest(ctx context.Context, id string) (*manifest.Manifest, error) {
	if m, ok := s.opt.Development[id]; ok {
		return m, nil
	}
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return nil, failure(err)
	}
	raw, err := os.ReadFile(filepath.Join(s.directory(in), "plugin.yaml"))
	if err != nil {
		return nil, failure(err)
	}
	parsed, err := manifest.Parse(raw)
	if err != nil {
		return nil, failure(err)
	}
	return parsed.Manifest, nil
}

// Package 返回组件的公开管理状态，省略实例配置与凭证。
type Package struct {
	ID, Role, AdapterID, Version, Source, PublisherKeyID, Pin string
	Enabled, Uninstalled, Signed                              bool
	Active, State                                             string
	Warnings                                                  []string `json:",omitempty"` // 仅安装结果携带清单警告。
	Pages                                                     []page   `json:"pages"`
}

// List 按安装顺序列出组件及其当前运行状态。
func (s *Service) List(ctx context.Context) ([]Package, error) {
	installs, err := s.opt.Store.Installs(ctx)
	if err != nil {
		return nil, failure(err)
	}
	result := make([]Package, 0, len(installs))
	for _, install := range installs {
		unlock := s.lock(install.PluginID)
		item, err := s.describe(ctx, install.PluginID)
		unlock()
		if err != nil {
			return nil, failure(err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) describe(ctx context.Context, id string) (Package, error) {
	in, err := s.opt.Store.GetInstall(ctx, id)
	if err != nil {
		return Package{}, err
	}
	enabled, err := s.opt.Store.Enabled(ctx, id)
	if err != nil {
		return Package{}, err
	}
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return Package{}, err
	}
	current := "Stopped"
	if active == "" && !in.Uninstalled {
		current = "Installed"
	}
	if enabled && !in.Uninstalled && in.Role == state.RoleWebUI && in.Entrypoint == "" {
		current = "Activated"
	} else if running := s.opt.State(id); running != "" {
		current = running
	} else if enabled && active != "" && !in.Uninstalled {
		current = "Degraded"
	}
	return Package{ID: id, Role: string(in.Role), AdapterID: in.AdapterID, Version: in.Version,
		Source: in.Source, PublisherKeyID: in.PublisherKeyID, Pin: in.Pin, Enabled: enabled,
		Uninstalled: in.Uninstalled, Signed: in.PublisherKeyID != "", Active: active, State: current, Pages: s.pages(id, nil)}, nil
}

func failure(err error) error {
	if err == nil {
		return nil
	}
	var confirm *ErrNeedsConfirm
	if errors.As(err, &confirm) {
		return err
	}
	var typed *validate.Error
	if errors.As(err, &typed) {
		return err
	}
	code := pb.ErrorCode_ERROR_CODE_INTERNAL
	var coded interface{ Code() pb.ErrorCode }
	if errors.As(err, &coded) {
		code = coded.Code()
	} else if errors.Is(err, state.ErrNotFound) {
		code = pb.ErrorCode_ERROR_CODE_NOT_FOUND
	}
	return &validate.Error{Code: code, Message: err.Error()}
}

func finish(err *error) { *err = failure(*err) }

func invalid(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: message}
}

func precondition(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION, Message: message}
}
func denied(message string) error {
	return &validate.Error{Code: pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, Message: message}
}
