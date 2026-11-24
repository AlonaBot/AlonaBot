//go:build unix

package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	genresources "github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

// afterSendPlugin 通过真实 SDK 检查回执独立返回、观察顺序和出站禁令。
func afterSendPlugin() *sdk.Plugin {
	p := sdk.NewPlugin()
	genresources.Bind(p, genresources.Resources{})
	p.Command("observe").Handle(func(c *sdk.Context) error {
		c.StopPropagation()
		ctx, cancel := context.WithTimeout(c, 100*time.Millisecond)
		result, err := sdk.Send(ctx, c.Session(), message.Text("original"), sdk.Via(pipelineInstanceID))
		cancel()
		if err != nil {
			return err
		}
		if err := pipelineAudit(pipelineRecord{Stage: "receipt-send", MessageID: result.PrimaryPlatformMessageID}); err != nil {
			return err
		}
		edited, err := c.Edit(result.Ref(), message.Text("edited"))
		if err != nil {
			return err
		}
		return pipelineAudit(pipelineRecord{Stage: "receipt-edit", MessageID: edited.PrimaryPlatformMessageID})
	})
	for _, id := range []string{"a", "b"} {
		p.AfterSend().ID(id).Handle(func(c *sdk.Context, _ string, result *message.SendResult, sendErr error) error {
			if sendErr != nil || result == nil || result.PrimaryPlatformMessageID == "" {
				return fmt.Errorf("incomplete platform receipt: %v %v", result, sendErr)
			}
			operation := "send"
			if c.Editing() != "" {
				operation = "edit"
			}
			if err := pipelineAudit(pipelineRecord{Stage: "observer-" + id, Action: operation, MessageID: result.PrimaryPlatformMessageID}); err != nil {
				return err
			}
			if id == "a" {
				request, err := http.NewRequestWithContext(c, http.MethodGet, os.Getenv(pipelineAuditEnv)+"/release", nil)
				if err != nil {
					return err
				}
				response, err := http.DefaultClient.Do(request)
				if err != nil {
					return err
				}
				if err := response.Body.Close(); err != nil {
					return err
				}
				if _, err := c.SendTo("loop:group#topic", message.Text("recursive"), sdk.Via(pipelineInstanceID)); errcode.Of(err) != errcode.FailedPrecondition {
					return fmt.Errorf("observer send ban: %v", err)
				}
				if _, err := c.Edit(result.Ref(), message.Text("recursive")); errcode.Of(err) != errcode.FailedPrecondition {
					return fmt.Errorf("observer edit ban: %v", err)
				}
			}
			return pipelineAudit(pipelineRecord{Stage: "observer-done", Route: id, Action: operation, MessageID: result.PrimaryPlatformMessageID})
		})
	}
	return p
}

// TestAfterSendReceipt 验证真实 Wire/SDK 回执先于阻塞观察返回，Send/Edit 均只执行一次。
func TestAfterSendReceipt(t *testing.T) {
	for _, stdio := range []bool{false, true} {
		t.Run(fmt.Sprint("stdio=", stdio), func(t *testing.T) {
			values := tune.Defaults()
			values.AfterSendConcurrency, values.AfterSendCapacity = 1, 2
			h := newPipelineAt(t, t.TempDir(), values, stdio)
			h.install("platform", "adapter")
			h.launch("platform", "adapter", 0)
			h.install("observers", "after-send")
			h.launch("observers", "after-send", 0)
			h.inject(pipelineInput{Text: "/observe", User: "alice"})
			h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "receipt-edit")) == 1 })
			h.idle()
			if done := pipelineStages(h.snapshot(), "observer-done"); len(done) != 0 {
				t.Fatalf("blocked observer unexpectedly completed: %v", done)
			}
			drainCtx, cancelDrain := context.WithTimeout(t.Context(), pipelineTimeout)
			defer cancelDrain()
			drained := make(chan error, 1)
			go func() { drained <- h.assembly.hooks.Close(drainCtx) }()
			// Send 的子请求已返回并取消；观察预算仍有效，下一批次等前一批释放执行槽。
			for count := 1; count <= 2; count++ {
				h.release <- struct{}{}
				h.await(func(records []pipelineRecord) bool { return len(pipelineStages(records, "observer-done")) >= 2*count })
			}
			if err := <-drained; err != nil {
				t.Fatalf("observer drain: %v", err)
			}
			var operations, ids []string
			order := map[string][]string{}
			for _, record := range h.snapshot() {
				switch record.Stage {
				case "send", "edit":
					operations = append(operations, record.Stage)
				case "observer-done":
					order[record.Action] = append(order[record.Action], record.Route)
					ids = append(ids, record.MessageID)
				}
			}
			if !reflect.DeepEqual(operations, []string{"send", "edit"}) || !reflect.DeepEqual(order["send"], []string{"a", "b"}) || !reflect.DeepEqual(order["edit"], []string{"a", "b"}) {
				t.Fatalf("platform=%v observers=%v", operations, order)
			}
			for _, id := range ids {
				if id != "1" {
					t.Fatalf("platform message ID changed: %v", ids)
				}
			}
		})
	}
}
