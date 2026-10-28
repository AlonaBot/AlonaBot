// Package main 运行通过环境描述命令的集成测试插件。
package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/AlonaBot/AlonaBot/sdk"
)

const commandsEnv = "ALONABOT_TEST_COMMANDS"

type command struct {
	Path     []string
	Aliases  [][]string
	Priority int32
}

type plugin struct {
	sdk.Base
	commands []command
}

// Commands 以完整路径注册环境中给出的命令。
func (p *plugin) Commands(r *sdk.CommandRegistry) {
	for _, c := range p.commands {
		options := []sdk.RouteOption{sdk.Priority(c.Priority)}
		for _, path := range c.Aliases {
			options = append(options, sdk.Alias(path...))
		}
		register(r, c.Path, options)
	}
}

func register(r *sdk.CommandRegistry, path []string, options []sdk.RouteOption) {
	if len(path) > 1 {
		r.Group(path[0], func(g *sdk.CommandGroup) { register(g, path[1:], options) })
		return
	}
	r.Command(strings.Join(path, ""), func(*sdk.Context) error { return nil }, options...)
}

func main() {
	p := &plugin{}
	if err := json.Unmarshal([]byte(os.Getenv(commandsEnv)), &p.commands); err != nil {
		slog.Error("decode test commands failed", "error", err)
		os.Exit(1)
	}
	if err := sdk.Run(p); err != nil {
		slog.Error("command plugin stopped with an error", "error", err)
		os.Exit(1)
	}
}
