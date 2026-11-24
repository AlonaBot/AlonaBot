package manage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

// 后台任务的状态。
const (
	taskRunning   = "running"
	taskSucceeded = "succeeded"
	taskFailed    = "failed"
)

// task 是一次后台管理操作。日志收集步骤说明与 go 命令的标准错误；mu 保护除 id、action、cancel 外的字段。
//
// shortcut: 任务只保存在内存中，Core 重启后丢失，也不会自动清理；需要跨重启查询或任务量变大时再持久化并设保留期。
type task struct {
	id, action string
	cancel     context.CancelFunc

	mu         sync.Mutex
	status     string
	log        bytes.Buffer
	result     any
	err        error
	createdAt  time.Time
	finishedAt time.Time
}

// Write 追加任务日志，供 go 命令与构建步骤并发写入。
func (t *task) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.log.Write(p)
}

// taskView 是任务的公开状态。失败时 HTTPStatus 与 Error 与同步端点的错误响应一致，
// 包括需要确认时的 409 与 {"affected":[…]}。
type taskView struct {
	ID         string     `json:"id"`
	Action     string     `json:"action"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Log        *string    `json:"log,omitempty"`
	Result     any        `json:"result,omitempty"`
	HTTPStatus int        `json:"http_status,omitempty"`
	Error      any        `json:"error,omitempty"`
}

func (t *task) view(withLog bool) taskView {
	t.mu.Lock()
	defer t.mu.Unlock()
	v := taskView{ID: t.id, Action: t.action, Status: t.status, CreatedAt: t.createdAt, Result: t.result}
	if !t.finishedAt.IsZero() {
		finished := t.finishedAt
		v.FinishedAt = &finished
	}
	if withLog {
		log := t.log.String()
		v.Log = &log
	}
	if t.err != nil {
		v.HTTPStatus, v.Error = errorResponse(t.err)
	}
	return v
}

// startTask 在后台运行 run 并立即返回任务 ID。Core 开始停机或任务被删除时取消 ctx；
// [Service.Wait] 等待全部任务结束。
func (s *Service) startTask(action string, run func(ctx context.Context, log io.Writer) (any, error)) string {
	ctx, cancel := context.WithCancel(context.Background())
	t := &task{id: rand.Text(), action: action, cancel: cancel, status: taskRunning, createdAt: time.Now()}
	s.tasks.Store(t.id, t)
	s.running.Go(func() {
		defer cancel()
		if s.opt.Stopping != nil {
			go func() { // 停机信号与任务结束谁先到都会让这个 goroutine 退出。
				select {
				case <-s.opt.Stopping:
					cancel()
				case <-ctx.Done():
				}
			}()
		}
		result, err := run(ctx, t)
		t.mu.Lock()
		defer t.mu.Unlock()
		t.finishedAt = time.Now()
		if err != nil {
			t.status, t.err = taskFailed, err
			slog.Warn(fmt.Sprintf("Management task %s failed during %s: %v", t.id, action, err), "task_id", t.id, "action", action, "error", err)
			return
		}
		t.status, t.result = taskSucceeded, result
	})
	return t.id
}

// Wait 等待全部后台任务结束；调用前应关闭 [Options.Stopping] 使任务取消。
func (s *Service) Wait() { s.running.Wait() }

func (s *Service) taskRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := make([]taskView, 0)
		s.tasks.Range(func(_, value any) bool {
			tasks = append(tasks, value.(*task).view(false))
			return true
		})
		slices.SortFunc(tasks, func(a, b taskView) int { return a.CreatedAt.Compare(b.CreatedAt) })
		w.Header().Set("Cache-Control", "no-store")
		respond(w, struct {
			Tasks []taskView `json:"tasks"`
		}{tasks}, nil)
	})
	mux.HandleFunc("GET /api/v1/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		value, found := s.tasks.Load(r.PathValue("id"))
		if !found {
			respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "task not found"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		respond(w, value.(*task).view(true), nil)
	})
	// DELETE 取消运行中的任务，或移除已结束的任务记录。
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		value, found := s.tasks.Load(r.PathValue("id"))
		if !found {
			respond(w, nil, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "task not found"})
			return
		}
		t := value.(*task)
		t.cancel()
		if t.view(false).Status != taskRunning {
			s.tasks.Delete(t.id)
		}
		respond(w, struct{}{}, nil)
	})
}

// respondTask 以 202 返回新任务的 ID。
func respondTask(w http.ResponseWriter, id string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(struct {
		TaskID string `json:"task_id"`
	}{id}); err != nil {
		slog.Error(fmt.Sprintf("Failed to write the management response: %v", err), "error", err)
	}
}
