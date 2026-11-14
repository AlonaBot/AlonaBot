//go:build unix

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestDevelopmentReload 首次贯通目录构建、真实插件/Adapter 进程与热部署；多列理由：现有生命周期测试仅覆盖安装进程。
func TestDevelopmentReload(t *testing.T) {
	plugin, adapter, dataDir := developmentFixture(t)
	c := startEmptyCore(t, dataDir, plugin, adapter, "--reload", "--args", "DEV_RELOAD_POLL=50ms")
	pluginPID := developmentPID(t, c, "hello")
	adapterPID := developmentPID(t, c, "loop")
	corePID := c.cmd.Process.Pid
	developmentReply(t, c, dataDir, "source:default")

	// 1. 源码成功替换只改变插件进程，编译失败保留刚才的可用版本。
	source := filepath.Join(plugin, "main.go")
	raw := developmentRead(t, source)
	developmentWrite(t, source, strings.ReplaceAll(string(raw), `"source:"`, `"changed:"`))
	developmentReply(t, c, dataDir, "changed:default")
	if developmentPID(t, c, "hello") == pluginPID {
		t.Fatal("plugin PID did not change")
	}
	if developmentPID(t, c, "loop") != adapterPID || c.cmd.Process.Pid != corePID {
		t.Fatal("unrelated process restarted")
	}
	developmentWrite(t, filepath.Join(plugin, "broken.go"), "package main\nfunc broken( {\n")
	c.waitFor(`(?s)development build failed.*broken.go`)
	developmentReply(t, c, dataDir, "changed:default")
	if err := os.Remove(filepath.Join(plugin, "broken.go")); err != nil {
		t.Fatal(err)
	}
	// 等待恢复构建，下一步的配置观察从稳定版本开始。
	deadline := time.Now().Add(20 * time.Second)
	previous := developmentPID(t, c, "hello")
	for developmentPID(t, c, "hello") == previous {
		if time.Now().After(deadline) {
			t.Fatalf("recovery build did not restart plugin:\n%s", c.log())
		}
		time.Sleep(20 * time.Millisecond)
	}
	developmentReply(t, c, dataDir, "changed:default")

	// 2. 只改配置保留二进制；Schema 的缺省值改动必须重新嵌入并启动。
	binary := filepath.Join(dataDir, "dev", "bin", "hello")
	before := developmentRead(t, binary)
	info, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	developmentWrite(t, filepath.Join(dataDir, "plugins", "hello", "config.json"), `{"text":"configured"}`)
	developmentReply(t, c, dataDir, "changed:configured")
	afterInfo, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, developmentRead(t, binary)) || !info.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("config-only change rebuilt executable")
	}
	developmentWrite(t, filepath.Join(dataDir, "plugins", "hello", "config.json"), `{}`)
	developmentReply(t, c, dataDir, "changed:default")
	developmentWrite(t, filepath.Join(plugin, "config_schema.json"), `{"text":{"type":"string","default":"schema"}}`)
	developmentReply(t, c, dataDir, "changed:schema")
	if bytes.Equal(before, developmentRead(t, binary)) {
		t.Fatal("schema change did not rebuild executable")
	}
	if developmentPID(t, c, "loop") != adapterPID {
		t.Fatal("schema change restarted Adapter")
	}
	if err := c.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if code := c.wait(15 * time.Second); code != 0 {
		t.Fatalf("Core exit %d:\n%s", code, c.log())
	}
}

func developmentFixture(t *testing.T) (string, string, string) {
	t.Helper()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	plugin, adapter, dataDir := t.TempDir(), t.TempDir(), t.TempDir()
	for _, project := range []struct{ dir, id, role, source string }{
		{plugin, "hello", "plugin", developmentPluginSource},
		{adapter, "loop", "adapter", developmentAdapterSource},
	} {
		developmentWrite(t, filepath.Join(project.dir, "go.mod"), fmt.Sprintf("module example.com/%s\n\ngo 1.27.1\n\nrequire github.com/AlonaBot/AlonaBot v0.0.0\nreplace github.com/AlonaBot/AlonaBot => %s\n", project.id, filepath.ToSlash(repository)))
		manifest := fmt.Sprintf("id: %s\nname: Development\nversion: 0.1.0\nauthor: AlonaBot\nsource: github.com/AlonaBot/AlonaBot\nrole: %s\n", project.id, project.role)
		if project.role == "adapter" {
			manifest += "adapter_id: loopback\n"
			developmentWrite(t, filepath.Join(project.dir, "adapter_config_schema.json"), `{"output":{"type":"string"}}`)
		}
		developmentWrite(t, filepath.Join(project.dir, "plugin.yaml"), manifest)
		developmentWrite(t, filepath.Join(project.dir, "main.go"), project.source)
		if project.role == "plugin" {
			developmentWrite(t, filepath.Join(project.dir, "config_schema.json"), `{"text":{"type":"string","default":"default"}}`)
		}
		command := exec.CommandContext(t.Context(), "go", "mod", "tidy")
		command.Dir = project.dir
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("tidy fixture: %v\n%s", err, output)
		}
	}
	developmentWrite(t, filepath.Join(dataDir, "core_config.json"), fmt.Sprintf(`{"listen":"127.0.0.1:0","adapters":{"loopback":{"instances":{"one":{"output":%q},"two":{"output":%q}}}}}`, filepath.Join(dataDir, "sent-one"), filepath.Join(dataDir, "sent-two")))
	return plugin, adapter, dataDir
}

func developmentWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func developmentRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func developmentPID(t *testing.T, c *core, id string) int {
	t.Helper()
	pattern := `plugin activated plugin_id=` + regexp.QuoteMeta(id) + ` generation_id=\S+ mode=\w+ pid=(\d+)`
	c.waitFor(pattern)
	matches := regexp.MustCompile(pattern).FindAllStringSubmatch(c.log(), -1)
	pid, err := strconv.Atoi(matches[len(matches)-1][1])
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

func developmentReply(t *testing.T, c *core, dir, want string) {
	t.Helper()
	if err := os.Remove(filepath.Join(dir, "sent-one")); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	token := strings.TrimSpace(string(developmentRead(t, filepath.Join(dir, "admin_token"))))
	deadline := time.Now().Add(20 * time.Second)
	for {
		request, err := http.NewRequest(http.MethodPost, "http://"+c.address()+"/plugins/loop/inject?instance=one", nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = "127.0.0.1:7650"
		request.Header.Set("Authorization", "Bearer "+token)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read injection response: %v %v", readErr, closeErr)
		}
		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("inject: %d %s\n%s", response.StatusCode, body, c.log())
		}
		time.Sleep(20 * time.Millisecond)
		raw, err := os.ReadFile(filepath.Join(dir, "sent-one"))
		if err == nil && string(raw) == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("reply = %q, want %q (read %v)\n%s", raw, want, err, c.log())
		}
	}
}

const developmentPluginSource = `package main
import (
 "github.com/AlonaBot/AlonaBot/sdk"
 "github.com/AlonaBot/AlonaBot/sdk/message"
)
func New() *sdk.Plugin {
 p := sdk.NewPlugin()
 p.Command("hello").Handle(func(c *sdk.Context) (message.Content, error) { return message.Chain{message.Text("source:"+c.Config().String("text"))}, nil })
 return p
}
`

const developmentAdapterSource = `package loopback
import (
 "context"
 "encoding/json"
 "fmt"
 "net/http"
 "os"
 "sync"
 "sync/atomic"
 "time"
 "github.com/AlonaBot/AlonaBot/sdk/adapter"
 "github.com/AlonaBot/AlonaBot/sdk/errcode"
 "github.com/AlonaBot/AlonaBot/sdk/message"
)
type Loopback struct { mu sync.Mutex; clients map[string]adapter.Core; sequence atomic.Uint64 }
func New() *Loopback { return &Loopback{clients: map[string]adapter.Core{}} }
func (a *Loopback) Descriptor() *adapter.Descriptor {
 return &adapter.Descriptor{AdapterID:"loopback", HTTP: []adapter.HTTPRoute{{Method:http.MethodPost,Path:"/inject",Handler:http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  id:=r.URL.Query().Get("instance"); a.mu.Lock(); core:=a.clients[id]; a.mu.Unlock()
  if core==nil { http.Error(w,"instance unavailable",http.StatusServiceUnavailable); return }
  number:=fmt.Sprint(a.sequence.Add(1))
  _,err:=core.Deliver(r.Context(), &message.Event{ID:number,Type:"message.created",Source:message.SourceAdapter,OccurredAt:time.Now(),AdapterID:"loopback",AdapterInstanceID:id,AccountID:"bot-"+id,Session:message.SID("loopback","chat",""),ChatType:message.ChatPrivate,Actor:&message.Actor{ExternalUserID:"user",AdapterInstanceID:id},Message:&message.PlatformMessage{PlatformMessageID:number,Chain:message.Chain{message.Text("/hello")}},Payload:&message.CreatedPayload{PlatformMessageID:number,Chain:message.Chain{message.Text("/hello")},ToMe:true}}, nil)
  if err!=nil { http.Error(w,err.Error(),http.StatusInternalServerError); return }; w.WriteHeader(http.StatusNoContent)
 })}}}
}
type instance struct { owner *Loopback; id, output string }
func (a *Loopback) Start(_ context.Context, req adapter.StartRequest, core adapter.Core) (adapter.Instance,error) {
 var config struct { Output string }; if err:=json.Unmarshal(req.ConfigJSON,&config);err!=nil{return nil,err}
 a.mu.Lock(); a.clients[req.InstanceID]=core; a.mu.Unlock()
 if err:=os.WriteFile(config.Output+".started",[]byte(fmt.Sprintf("%d",time.Now().UnixNano())),0600);err!=nil{return nil,err}
 return &instance{owner:a,id:req.InstanceID,output:config.Output},nil
}
func (i *instance) Account() adapter.Account {return adapter.Account{ID:"bot-"+i.id}}
func (*instance) Capabilities() []message.Capability {return nil}
func (i *instance) Stop(context.Context) error {i.owner.mu.Lock();delete(i.owner.clients,i.id);i.owner.mu.Unlock();return nil}
func (i *instance) Send(_ context.Context,to string,chain message.Chain)(*message.SendResult,*errcode.Error){
 var text string;for _,segment:=range chain {if value,ok:=segment.(message.TextSegment);ok{text+=value.Text}}
 if err:=os.WriteFile(i.output,[]byte(text),0600);err!=nil{return nil,&errcode.Error{Code:errcode.Internal,Message:err.Error()}}
 return &message.SendResult{Status:message.SendAccepted,Session:to,AdapterInstanceID:i.id,PrimaryPlatformMessageID:"sent",PlatformMessageIDs:[]string{"sent"},PartResults:[]message.PartResult{{PartIndex:0,SegmentIndexes:[]uint32{0},Status:message.PartAccepted,PlatformMessageIDs:[]string{"sent"}}}},nil
}
`
