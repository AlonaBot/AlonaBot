package app

import (
	"fmt"
	"log/slog"
)

func logActivation(id, generation, mode string, pid int) {
	slog.Info(fmt.Sprintf("plugin activated: %s (generation %s, %s, PID %d)", id, generation, mode, pid), "plugin_id", id, "generation_id", generation, "mode", mode, "pid", pid)
}
