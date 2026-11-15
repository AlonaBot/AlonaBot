// Package registry 保存根 SDK 与分类子包共用的注册状态和调用边界。
//
// Scope 经私有嵌入只允许插件与子插件登记；状态归插件对象，按原登记顺序收集。
// 根包设置运行时入口，子包复用 Session 门禁、父调用和资源身份（CONTRACTS §3、§9.1）。
package registry

import (
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// Scope 关联一个插件或子插件与其共享注册器。
type Scope[C any, P interface{ Names() []string }] struct {
	Plugin        any
	SubID         string
	Services      *Services[C, P]
	Subscriptions *Subscriptions[C]
	Tasks         *Tasks[C]
	TaskContext   func(C) (*pb.Event, *wire.Inbound)
	Check         func()
	Button        func(string, uint32) (any, func(func(string) (any, error), func(C, any) error))
}

// State 通过私有嵌入提供封闭登记方法，作用域字段留在内部包。
type State[C any, P interface{ Names() []string }] struct{ scope Scope[C, P] }

func (s *State[C, P]) registrationScope() *Scope[C, P] { return &s.scope }

// NewState 将已装配作用域交给插件对象持有。
func NewState[C any, P interface{ Names() []string }](scope Scope[C, P]) State[C, P] {
	return State[C, P]{scope: scope}
}

// Owner 封闭登记入口，由 sdk 的插件和子插件通过私有嵌入实现。
type Owner[C any, P interface{ Names() []string }] interface{ registrationScope() *Scope[C, P] }

// For 返回登记对象持有的作用域。
func For[C any, P interface{ Names() []string }](owner Owner[C, P]) *Scope[C, P] {
	return owner.registrationScope()
}

// Entry 返回插件构造时记录的入口包。
func (s *Scope[C, P]) Entry() string { pkg, _ := generated.Entry(s.Plugin); return pkg }

// Resources 按对象与入口包读取同一份生成资源。
func (s *Scope[C, P]) Resources() generated.Resources {
	r, _ := generated.Lookup(s.Plugin, s.Entry())
	return r
}

// Services 保存按登记顺序构建的服务和 Schema。
type Services[C any, P interface{ Names() []string }] struct {
	Services []Service[C]
	Builders []*ServiceBuilder[C, P]
	Schemas  map[string][]byte
	Errs     []error
}

// Service 保存冻结的服务描述和处理函数。
type Service[C any] struct {
	Plugin     any
	SubID      string
	Descriptor *pb.RouteDescriptor
	Invoke     func(C, []byte) ([]byte, error)
}

// Access 保留服务访问规则的内部枚举。
type Access int

const (
	AccessPrivate Access = iota
	AccessAllowlist
	AccessPublic
)

// ServiceBuilder 保存收集前可修改的服务声明。
type ServiceBuilder[C any, P interface{ Names() []string }] struct {
	Scope        *Scope[C, P]
	Name, Target string
	Encoding     pb.PayloadEncoding
	Access       Access
	Allowlist    []string
	Permission   P
	Handles      int
	AfterHandle  bool
	Prepare      func() (Service[C], error)
}

// Route 固定服务的 Schema、权限和访问规则。
func (b *ServiceBuilder[C, P]) Route(inputID, outputID string, inputVersion, outputVersion uint32, encoding pb.PayloadEncoding, invoke func(C, []byte) ([]byte, error)) Service[C] {
	access := pb.ServiceAccess_SERVICE_ACCESS_UNSPECIFIED
	switch b.Access {
	case AccessPrivate:
		access = pb.ServiceAccess_SERVICE_ACCESS_PRIVATE
	case AccessAllowlist:
		access = pb.ServiceAccess_SERVICE_ACCESS_ALLOWLIST
	case AccessPublic:
		access = pb.ServiceAccess_SERVICE_ACCESS_PUBLIC
	}
	return Service[C]{Plugin: b.Scope.Plugin, SubID: b.Scope.SubID, Descriptor: &pb.RouteDescriptor{
		RouteId: "svc." + b.Name, Kind: pb.RouteKind_ROUTE_KIND_SERVICE, DefaultEnabled: true,
		RequiredPermissions: b.Permission.Names(),
		InputSchemaId:       inputID, InputSchemaVersion: inputVersion, OutputSchemaId: outputID, OutputSchemaVersion: outputVersion,
		Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: b.Name, Access: access, AllowlistPluginIds: b.Allowlist, Encoding: encoding}},
	}, Invoke: invoke}
}

// Subscriptions 保存订阅构建链与冻结的回调。
type Subscriptions[C any] struct {
	Pending       []func() error
	Subscriptions []Subscription[C]
	Errs          []error
}

// Subscription 保存订阅描述及无用户权限的回调。
type Subscription[C any] struct {
	Plugin     any
	SubID      string
	Descriptor *pb.RouteDescriptor
	Invoke     func(C, []byte) error
}

// Tasks 保存按登记顺序完成的任务构建链。
type Tasks[C any] struct{ Pending []func() (Task[C], error) }

// Task 保存任务名、子插件归属及冻结的处理函数。
type Task[C any] struct {
	Name, SubID string
	Invoke      func(C) error
}

// Descriptor 返回任务的固定路由描述。
func (t Task[C]) Descriptor() *pb.RouteDescriptor {
	return &pb.RouteDescriptor{RouteId: "task." + t.Name, Kind: pb.RouteKind_ROUTE_KIND_TASK, DefaultEnabled: true, Spec: &pb.RouteDescriptor_Task{Task: &pb.TaskSpec{Name: t.Name}}}
}
