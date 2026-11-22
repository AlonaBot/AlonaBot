package sdk

import (
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// httpRuntime 让 Session 只在作者登记 HTTP 路由时链接 Listener 实现。
type httpRuntime interface {
	start(*pb.Bootstrap) error
	report(*pb.Register)
	registrations() []httpRoute
	close(time.Time)
	activate(*activation)
}

func (l launch) startHTTP(boot *pb.Bootstrap) error {
	if l.http == nil {
		return nil
	}
	return l.http.start(boot)
}

func (l launch) reportHTTP(reg *pb.Register) {
	if l.http != nil {
		l.http.report(reg)
	}
}

func (l launch) closeHTTP(deadline time.Time) {
	if l.http != nil {
		l.http.close(deadline)
	}
}

func (r *httpRegistry) newRuntime(routes []httpRoute, runtime *componentSession) httpRuntime {
	if r.create == nil {
		return nil
	}
	return r.create(routes, runtime)
}

func (r *httpRegistry) collect() {
	if r.collectNow == nil {
		r.routes, r.errs = nil, nil
		return
	}
	r.collectNow()
}
