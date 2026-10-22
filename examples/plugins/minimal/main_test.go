// 本文件验证 minimal 示例的声明能被 SDK 收集，配置回调与迁移签名成立，Run 对未支持的声明返回 UNSUPPORTED。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/config"
	"github.com/AlonaBot/AlonaBot/sdk/gormdb"
)

// 示例的各个声明能被 SDK 组合与收集，证明 SDK 的公共签名成立。
func TestDemoComposes(t *testing.T) {
	root := &Demo{}
	reg, err := sdk.Collect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Commands) != 2 || len(reg.HTTP) != 1 || len(reg.Schedules) != 1 {
		t.Fatalf("收集结果不符: %+v", reg)
	}
	if reg.Commands[0].Mode != pb.HandlerMode_HANDLER_MODE_AUTO_REPLY || reg.Commands[1].Mode != pb.HandlerMode_HANDLER_MODE_MANUAL {
		t.Fatalf("Handler 模式识别错误: %v / %v", reg.Commands[0].Mode, reg.Commands[1].Mode)
	}
	if got := reg.Commands[1].Params; len(got) != 2 || got[0] != pb.ParamType_PARAM_TYPE_STRING || got[1] != pb.ParamType_PARAM_TYPE_INT {
		t.Fatalf("参数类型识别错误: %v", got)
	}

	cfg, ok := any(root).(config.Provider)
	if !ok {
		t.Fatal("Demo 应实现 config.Provider")
	}
	if _, ok := cfg.Config().(config.Loader); !ok {
		t.Fatal("Config 应实现 config.Loader")
	}
	if _, ok := cfg.Config().(config.Validator); !ok {
		t.Fatal("Config 应实现 config.Validator")
	}
	if _, ok := any(root).(sdk.Starter); !ok {
		t.Fatal("Demo 应实现 sdk.Starter")
	}
	if mode, err := gormdb.CheckMigrators(root); err != nil || mode != gormdb.ModeNone {
		t.Fatalf("示例没有迁移，应为 none: %v %v", mode, err)
	}

	out, err := (&Config{}).OnLoad(context.Background(), 1, json.RawMessage(`{"hello":"hi","keep":1}`))
	if err != nil || string(out) != `{"greeting":"hi","keep":1}` {
		t.Fatalf("OnLoad 结果 %s, %v", out, err)
	}
}

// 运行时尚未实现：Run 返回明确的 UNSUPPORTED，而不是 panic 或假装成功。
func TestRunNotImplemented(t *testing.T) {
	// Core 启动插件时注入 runtime.lock 的绝对路径；示例只依赖 SDK，故直接写环境变量名。
	t.Setenv("ALONABOT_RUNTIME_LOCK", filepath.Join(t.TempDir(), "runtime.lock"))
	var e *sdk.Error
	if err := sdk.Run(&Demo{}); !errors.As(err, &e) || e.Code != pb.ErrorCode_ERROR_CODE_UNSUPPORTED {
		t.Fatalf("Run 应返回 UNSUPPORTED，实际 %v", err)
	}
}
