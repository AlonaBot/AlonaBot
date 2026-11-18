package sdk

import (
	"errors"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/internal/registry"
)

const taskRoutePrefix = "task."

type taskRegistry = registry.Tasks[*Context]
type taskRoute = registry.Task[*Context]

func collectTasks(root *Plugin, reg *registration) error {
	reg.tasks = nil
	var errs []error
	for _, finish := range root.tasks.Pending {
		task, err := finish()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		reg.tasks = append(reg.tasks, task)
		reg.subPlugins[taskRoutePrefix+task.Name] = task.SubID
	}
	// 仅校验 task 前缀相关路由；HTTP Listener 在后续准备阶段才创建。
	registration := &pb.Register{CandidateConfigJson: "{}"}
	for _, route := range buildRegister(reg, "{}", nil).Routes {
		if strings.HasPrefix(route.RouteId, taskRoutePrefix) {
			registration.Routes = append(registration.Routes, route)
		}
	}
	errs = append(errs, validate.Register(registration))
	return errors.Join(errs...)
}
