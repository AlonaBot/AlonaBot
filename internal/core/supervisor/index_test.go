//go:build unix

package supervisor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/internal/core/host"
	"github.com/AlonaBot/AlonaBot/internal/core/router"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

const commandTestEnv = "ALONABOT_TEST_COMMANDS"
const commandConflictFault = "command-path-alias-conflict"

func commandSpec(t *testing.T, binary, id, commands string) (*state.Store, Spec) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	store, err := state.Open(ctx, filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	// 固定 fixture 摘要只用于测试握手双方的一致性。
	digest := make([]byte, 32)
	if err := store.AddInstall(ctx, state.Install{Role: state.RolePlugin, PluginID: id, Source: "test", Entrypoint: binary, BuildDigest: digest}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddGeneration(ctx, id, "gen-1", "{}", 1); err != nil {
		t.Fatal(err)
	}
	pluginDir := filepath.Join(dir, "plugins", id)
	if err := os.MkdirAll(pluginDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return store, Spec{Role: state.RolePlugin, PluginID: id, PluginDir: pluginDir, Executable: binary, BuildDigest: digest, GenerationID: "gen-1",
		Bootstrap:    &pb.Bootstrap{GenerationId: "gen-1", CandidateConfigJson: "{}", TargetConfigVersion: 1, DatabaseDescriptor: &pb.DatabaseDescriptor{Backend: pb.DatabaseBackend_DATABASE_BACKEND_NONE, MigrationMode: pb.MigrationMode_MIGRATION_MODE_NONE}},
		NewCandidate: true, Limits: fastLimits, Stop: fastStop, Log: &logBuf{}, ExtraEnv: map[string]string{commandTestEnv: commands}}
}

func conflictingCommands() *pb.Register {
	return &pb.Register{CandidateConfigJson: "{}", Routes: []*pb.RouteDescriptor{
		{RouteId: "cmd.rss", Kind: pb.RouteKind_ROUTE_KIND_COMMAND, Spec: &pb.RouteDescriptor_Command{Command: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH, Path: []string{"rss"}}}},
		{RouteId: "cmd.other", Kind: pb.RouteKind_ROUTE_KIND_COMMAND, Spec: &pb.RouteDescriptor_Command{Command: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH, Path: []string{"other"}, Aliases: []*pb.CommandAlias{{Path: []string{"rss"}}}}}},
	}}
}

// TestCommandIndexProcesses 验证真实插件注册后的稳定索引与失败 REGISTER 的原子性。
func TestCommandIndexProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "cmdplugin")
	if output, err := exec.Command("go", "build", "-o", binary, "./testdata/cmdplugin").CombinedOutput(); err != nil {
		t.Fatalf("build command plugin: %v\n%s", err, output)
	}
	definitions := map[string]string{
		"a": `[{"Path":["rss"]},{"Path":["rss","add"]}]`,
		"b": `[{"Path":["rss"],"Aliases":[["feed"]]}]`,
		"c": `[{"Path":["rss"],"Priority":-1}]`,
	}
	for _, stdio := range []bool{false, true} {
		for _, order := range [][]string{{"a", "b", "c"}, {"a", "c", "b"}, {"b", "a", "c"}, {"b", "c", "a"}, {"c", "a", "b"}, {"c", "b", "a"}} {
			t.Run(fmt.Sprintf("stdio=%v/order=%v", stdio, order), func(t *testing.T) {
				previous := slog.Default()
				slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
				t.Cleanup(func() { slog.SetDefault(previous) })
				r := router.New(1, nil, nil, nil, nil)
				defer r.Close()
				for _, id := range order {
					store, spec := commandSpec(t, binary, id, definitions[id])
					spec.Stdio = stdio
					caller := router.NewCaller(spec.PluginID, spec.GenerationID)
					spec.Handler = router.NewMux().For(caller)
					running, err := Launch(context.Background(), store, spec)
					if err != nil {
						t.Fatalf("launch %s: %v; logs: %s", id, err, spec.Log.(*logBuf).String())
					}
					t.Cleanup(func() {
						result := running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
						if result.Err != nil || result.Killed {
							t.Errorf("stop %s: %+v", id, result)
						}
					})
					r.Attach(router.FromInstance(running.Instance), caller)
				}
				x := r.Commands()
				for _, c := range []struct {
					tokens        []string
					plugin, route string
					consumed      int
					alias         bool
				}{
					{[]string{"rss"}, "c", "cmd.rss", 1, false},
					{[]string{"feed"}, "b", "cmd.rss", 1, true},
					{[]string{"rss", "add", "x"}, "a", "cmd.rss.add", 2, false},
				} {
					entry, n, ok := x.Lookup(c.tokens)
					if !ok || entry.PluginID != c.plugin || entry.RouteID != c.route || n != c.consumed || entry.Alias != c.alias {
						t.Fatalf("Lookup(%v) = %+v, %d, %v", c.tokens, entry, n, ok)
					}
				}
				if _, _, ok := x.Lookup([]string{"rssx"}); ok {
					t.Fatal("rssx matched rss")
				}
				shadows := x.Shadows()
				if len(shadows) != 1 || !reflect.DeepEqual(shadows[0].Path, []string{"rss"}) || shadows[0].Winner.PluginID != "c" || len(shadows[0].Shadowed) != 2 || shadows[0].Shadowed[0].PluginID != "a" || shadows[0].Shadowed[1].PluginID != "b" {
					t.Fatalf("shadows = %+v", shadows)
				}
				r.Detach("c")
				entry, _, ok := r.Commands().Lookup([]string{"rss"})
				if !ok || entry.PluginID != "a" {
					t.Fatalf("after detach = %+v, %v", entry, ok)
				}

				// 原始插件绕过 SDK，必须在 REGISTER 阶段整体失败，且已发布索引不变。
				before := r.Commands()
				store, spec, _ := rawSpec(t, commandConflictFault, stdio)
				running, err := Launch(context.Background(), store, spec)
				if running != nil {
					running.Stop(time.Now(), pb.StopReason_STOP_REASON_DISABLE)
					t.Fatal("conflicting registration activated")
				}
				var phase *host.PhaseError
				if !errors.As(err, &phase) || phase.Phase != pb.SessionPhase_SESSION_PHASE_REGISTER || phase.Code != pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS {
					t.Fatalf("Launch conflict = %v; want REGISTER ALREADY_EXISTS", err)
				}
				if r.Commands() != before || !reflect.DeepEqual(r.Commands().Shadows(), before.Shadows()) {
					t.Fatal("failed registration changed command index")
				}
				if _, _, ok := r.Commands().Lookup([]string{"other"}); ok {
					t.Fatal("failed registration left a route")
				}
			})
		}
	}
}
