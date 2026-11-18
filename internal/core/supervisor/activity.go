package supervisor

import (
	"sync"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

type activitySession struct {
	running       *Running
	ready         bool
	stopping      bool
	idleStop      bool
	stopCancelled bool
	stopCommitted bool // 在 STOPPING 的不可撤销门禁处设置，由 Waker.mu 保护。
	references    int
	endedAt       time.Time
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
	p.coldStarts = append(p.coldStarts, at)
}

func (w *Waker) register(r *Running) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	p.current = &activitySession{running: r, endedAt: time.Now()}
	p.failure = nil
	p.notify()
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
	if current.stopping && current.idleStop && !current.stopCommitted {
		if current.active() || current.running.Instance.State() == host.StateDraining {
			current.stopCancelled = true
			return w.acquire(current), nil
		}
	}
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
				p := w.plugin(s.running.pluginID)
				if p.current == s {
					p.lastActivityAt = s.endedAt
				}
			}
		})
	}
}

func (w *Waker) stopping(r *Running, idle bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	if p.current == nil || p.current.running != r {
		return true
	}
	if idle && p.current.references != 0 {
		return false
	}
	p.current.stopping, p.current.idleStop = true, idle
	p.current.stopCancelled, p.current.stopCommitted = false, false
	p.notify()
	return true
}

// commitStop 与请求接纳共用门禁；成功后新请求只等待旧树退出。
func (w *Waker) commitStop(r *Running) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := w.plugin(r.pluginID).current
	if s == nil || s.running != r {
		return true
	}
	if s.idleStop && (s.stopCancelled || s.references != 0) {
		return false
	}
	s.stopCommitted = true
	return true
}

func (w *Waker) resume(r *Running) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	if p.current != nil && p.current.running == r {
		p.current.stopping, p.current.idleStop = false, false
		p.notify()
	}
}

func (w *Waker) released(r *Running) {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.plugin(r.pluginID)
	if p.current != nil && p.current.running == r {
		p.lastActivityAt = time.Now()
		p.current = nil
		p.notify()
	}
}
