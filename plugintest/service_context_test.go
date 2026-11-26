package plugintest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/AlonaBot/AlonaBot/internal/generated"
	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/service"
)

type serviceContextInfo struct {
	ChatType message.ChatType `json:"chat_type"`
	User     string           `json:"user"`
	Session  string           `json:"session"`
	HasEvent bool             `json:"has_event"`
}

// TestServiceInvocationContext 验证群聊命令经真实 Core 向另一插件的 JSON Service 透传会话身份。
func TestServiceInvocationContext(t *testing.T) {
	caller, provider := sdk.NewPlugin(), sdk.NewPlugin()
	typ := reflect.TypeFor[serviceContextInfo]()
	for p, id := range map[*sdk.Plugin]string{caller: "context-caller", provider: "context-provider"} {
		generated.Bind(p, generated.Resources{
			Manifest: []byte(fmt.Sprintf("id: %s\nname: %s\nversion: 0.1.0\nauthor: AlonaBot\nrole: plugin\nsource: github.com/AlonaBot/AlonaBot\n", id, id)),
			ServiceSchemas: map[string][]byte{
				typ.PkgPath() + "." + typ.Name(): []byte(`{"type":"object","properties":{"chat_type":{"type":"integer"},"user":{"type":"string"},"session":{"type":"string"},"has_event":{"type":"boolean"}},"required":["chat_type","user","session","has_event"]}`),
			},
		})
	}
	declaration := service.JSON[serviceContextInfo, serviceContextInfo]("context-provider/conversation.reset")
	service.Serve(provider, declaration).Access(service.AccessPublic).Handle(func(c *sdk.Context, _ serviceContextInfo) (serviceContextInfo, error) {
		return serviceContextInfo{ChatType: c.ChatType(), User: c.User(), Session: c.Session(), HasEvent: c.Event() != nil}, nil
	})
	caller.Command("new").Handle(func(c *sdk.Context) (message.Content, error) {
		got, err := declaration.Call(c, serviceContextInfo{})
		if err != nil {
			return nil, err
		}
		want := serviceContextInfo{ChatType: c.ChatType(), User: c.User(), Session: c.Session()}
		if got != want || got.ChatType != message.ChatGroup {
			return nil, fmt.Errorf("service context = %+v, want %+v with group chat type", got, want)
		}
		return message.Text("session reset"), nil
	})
	h := Start(t, Plugin(caller), Plugin(provider))
	h.Text(Group("room"), User{ID: "alice"}, "/new").RequireReply(t, "session reset")
}
