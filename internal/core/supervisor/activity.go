package supervisor

import (
	"slices"
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type activitySession struct {
	running    *Running
	ready      bool
	stopping   bool
	idleStop   bool
	references int
	endedAt    time.Time
}

// ActivityStatus 返回当前 Session 的活动快照，冷启动时刻由调用方独立持有。
type ActivityStatus struct {
	Session           *wire.Session
	Ready             bool // 当前 Session 完整绑定、存活且接纳新工作。
	References        int
	LastActivityAt    time.Time   // 当前 Session 最近一次引用归零，初值为登记时刻。
	ColdStarts        []time.Time // 尚未由空闲控制器确认消费的实际启动时刻。
	ColdStartSequence uint64      // ColdStarts 末项的启动序号；切片为空时仍保留总进度。
}

// Activity 在活动门禁内读取当前 Session，复制实际冷启动时刻。
func (w *Waker) Activity(pluginID string) ActivityStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(pluginID)
	status := ActivityStatus{ColdStarts: slices.Clone(p.coldStarts), ColdStartSequence: p.coldStartSequence}
	if s := p.current; s != nil {
		status.Session = s.running.Instance.Session
		status.Ready = s.ready && !s.stopping && s.active()
		status.References, status.LastActivityAt = s.references, s.endedAt
	}
	return status
}

func (s *activitySession) active() bool {
	if s.running.Instance.State() != host.StateActivated {
		return false
	}
	select {
	case <-s.running.Instance.Session.Done():
		return false
	default:
		return true
	}
}

func (w *Waker) started(id string, at time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(id)
	p.coldStartSequence++
	p.coldStarts = append(p.coldStarts, at)
}

// AcknowledgeColdStarts 回收控制器已消费的快照前缀，保留快照之后发生的启动。
// through 必须来自该插件的 Activity 快照，多个读者取得快照不会消耗记录。
func (w *Waker) AcknowledgeColdStarts(pluginID string, through uint64) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(pluginID)
	first := p.coldStartSequence - uint64(len(p.coldStarts))
	if through > first {
		p.coldStarts = slices.Clone(p.coldStarts[through-first:])
	}
}

func (w *Waker) register(r *Running, activity *activitySession) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	activity.running = r
	if activity.endedAt.IsZero() {
		activity.endedAt = time.Now()
	}
	p.current = activity
	p.failure = nil
	p.notify()
}

// requestActivity 在 wire 同步接纳时占用活动；提前于异步 Handler 和装配等待。
func (w *Waker) requestActivity(s *activitySession, in *wire.Inbound) (func(), *pb.ErrorInfo) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if s.stopping {
		phase := s.running.Instance.State()
		if in.ParentCallID == 0 || phase != host.StateActivated && phase != host.StateDraining {
			return nil, &pb.ErrorInfo{Code: pb.ErrorCode_ERROR_CODE_UNAVAILABLE, SafeMessage: "plugin Session is no longer accepting requests"}
		}
	}
	return w.acquire(s), nil
}

// Ready 在装配方完成当前 Session 的全部绑定后开放等待者。
func (w *Waker) Ready(running *Running) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(running.pluginID)
	if p.current != nil && p.current.running == running && !p.current.stopping {
		p.current.ready = true
		p.notify()
	}
}

// AcquireActivity 为当前匹配且仍接纳工作的 Session 取得引用，始终不启动进程。
func (w *Waker) AcquireActivity(pluginID string, session *wire.Session) (func(), error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(pluginID)
	if p.current == nil || p.current.running.Instance.Session != session {
		return nil, wakeError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session is no longer active")
	}
	current := p.current
	acceptedDrain := current.stopping && !current.idleStop && current.running.Instance.State() == host.StateDraining
	if !acceptedDrain && (current.stopping || !current.active()) {
		return nil, wakeError(pb.ErrorCode_ERROR_CODE_UNAVAILABLE, "plugin Session is no longer accepting activity")
	}
	return w.acquire(p.current), nil
}

// acquire 要求持有 w.mu；释放闭包幂等，且永远只扣减这次 Session 的计数。
func (w *Waker) acquire(s *activitySession) func() {
	s.references++
	var once sync.Once
	return func() {
		once.Do(func() {
			w.mu.Lock()
			defer w.mu.Unlock()
			s.references--
			if s.references == 0 {
				s.endedAt = time.Now()
			}
		})
	}
}

func (w *Waker) stopping(r *Running, idle bool, lastActivityAt *time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	if p.current == nil || p.current.running != r {
		return !idle
	}
	s := p.current
	if idle && (s.stopping || !s.ready || !s.active() || s.references != 0 || lastActivityAt != nil && !s.endedAt.Equal(*lastActivityAt)) {
		return false
	}
	// 此处与新活动互斥；一旦交接停止权，新请求等待下一次 Session。
	s.stopping, s.idleStop = true, idle
	p.notify()
	return true
}

func (w *Waker) released(r *Running) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	if p.current != nil && p.current.running == r {
		p.current = nil
		p.notify()
	}
}
