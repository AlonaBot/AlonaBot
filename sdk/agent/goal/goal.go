// Package goal 声明 Agent 持久目标的管理、验收与恢复契约。
package goal

import (
	"context"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/agent"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/llm"
	"github.com/AlonaBot/AlonaBot/sdk/service"
	"time"
)

// RecoveryPolicy 指定重启后的恢复方式。
type RecoveryPolicy string

const (
	// Manual 重启后暂停未完成目标，等待显式 Resume；零值采用本策略。
	Manual RecoveryPolicy = "manual"
	// Automatic 预留仅恢复重启前 Running 目标的策略。
	Automatic RecoveryPolicy = "automatic"
)

// Status 描述目标状态，Completed、Failed 和 Cancelled 为终态。
type Status string

const (
	// Pending 等待安排首个 Run。
	Pending Status = "pending"
	// Running 执行 Run 或独立评估。
	Running Status = "running"
	// Paused 保存进度并等待恢复。
	Paused Status = "paused"
	// Completed 已满足条件。
	Completed Status = "completed"
	// Failed 条件已被判定无法完成。
	Failed Status = "failed"
	// Cancelled 已被显式取消。
	Cancelled Status = "cancelled"
)

// Verdict 描述确定性 Checker 或独立模型的验收结果。
type Verdict string

const (
	// NotMet 继续运行，Reason 作为下一轮指导。
	NotMet Verdict = "not_met"
	// Met 条件已达成。
	Met Verdict = "met"
	// Impossible 依据现有证据无法完成。
	Impossible Verdict = "impossible"
)

// Evaluation 保存验收结论和完整原因。
type Evaluation struct {
	Verdict Verdict `json:"verdict"`
	Reason  string  `json:"reason"`
}

// CreateRequest 指定可验证条件和初始 Run；空 EvaluatorModel 使用本次 Run 选定模型。
type CreateRequest struct {
	Condition      string           `json:"condition"`
	Run            agent.RunRequest `json:"run"`
	EvaluatorModel string           `json:"evaluator_model,omitempty"`
	Checker        string           `json:"checker,omitempty"`
	RecoveryPolicy RecoveryPolicy   `json:"recovery_policy,omitempty"`
}

// Goal 返回持久状态及可信创建者；时间编码为 RFC3339。
type Goal struct {
	ID               string         `json:"id"`
	CreatorPluginID  string         `json:"creator_plugin_id"`
	Condition        string         `json:"condition"`
	RecoveryPolicy   RecoveryPolicy `json:"recovery_policy"`
	Status           Status         `json:"status"`
	RunCount         uint64         `json:"run_count"`
	LatestEvaluation *Evaluation    `json:"latest_evaluation,omitempty"`
	Reason           string         `json:"reason"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// IDRequest 携带待管理目标的 ID，归属由 Agent 根据可信调用身份检查。
type IDRequest struct {
	ID string `json:"id"`
}

// ListRequest 查询当前创建者的目标。
type ListRequest struct{}

// ListResponse 返回当前创建者的全部目标。
type ListResponse struct {
	Goals []Goal `json:"goals"`
}

// CheckerInput 携带条件与已保存的真实执行记录。
type CheckerInput struct {
	GoalID    string            `json:"goal_id"`
	Condition string            `json:"condition"`
	Runs      []agent.RunResult `json:"runs"`
}

// CreateService 绑定目标创建服务。
var CreateService = service.JSON[CreateRequest, Goal]("@agent/goal.create")

// GetService 绑定目标查询服务。
var GetService = service.JSON[IDRequest, Goal]("@agent/goal.get")

// ListService 绑定目标列表服务。
var ListService = service.JSON[ListRequest, ListResponse]("@agent/goal.list")

// PauseService 绑定协作暂停服务。
var PauseService = service.JSON[IDRequest, Goal]("@agent/goal.pause")

// ResumeService 绑定显式恢复服务。
var ResumeService = service.JSON[IDRequest, Goal]("@agent/goal.resume")

// CancelService 绑定终态取消服务。
var CancelService = service.JSON[IDRequest, Goal]("@agent/goal.cancel")

// Create 请求 Agent 持久化并安排目标。
func Create(ctx context.Context, req CreateRequest) (Goal, error) {
	if req.Condition == "" {
		return Goal{}, errcode.InvalidArgument.New("condition is required")
	}
	if req.RecoveryPolicy != "" && req.RecoveryPolicy != Manual && req.RecoveryPolicy != Automatic {
		return Goal{}, errcode.InvalidArgument.New("unknown recovery policy")
	}
	if err := llm.Validate(req.Run.Input); err != nil {
		return Goal{}, err
	}
	return CreateService.Call(ctx, req)
}

// Get 读取目标状态。
func Get(ctx context.Context, id string) (Goal, error) {
	return GetService.Call(ctx, IDRequest{ID: id})
}

// List 返回创建者的全部目标。
func List(ctx context.Context) ([]Goal, error) {
	out, err := ListService.Call(ctx, ListRequest{})
	return out.Goals, err
}

// Pause 协作取消当前 Run 并保存已确认步骤。
func Pause(ctx context.Context, id string) (Goal, error) {
	return PauseService.Call(ctx, IDRequest{ID: id})
}

// Resume 从持久记录安排新的 Run。
func Resume(ctx context.Context, id string) (Goal, error) {
	return ResumeService.Call(ctx, IDRequest{ID: id})
}

// Cancel 将目标置为 Cancelled 并协作取消当前工作。
func Cancel(ctx context.Context, id string) (Goal, error) {
	return CancelService.Call(ctx, IDRequest{ID: id})
}

// Checker 注册当前 Agent 可访问、允许受控重入的确定性验收服务。
func Checker(owner sdk.Registrar, id string) *service.Builder[CheckerInput, Evaluation] {
	target := "/goal.checker." + id
	if id == "" {
		target = "/"
	}
	return service.Serve(owner, service.JSON[CheckerInput, Evaluation](target)).Access(service.AccessAllowlist, "@agent").Reentrant()
}
