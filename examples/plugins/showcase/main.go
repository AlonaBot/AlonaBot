// showcase 演示 SDK 全部公共签名的组合方式；可以真实运行的最小插件见 examples/plugins/lifecycle。
//
// 示例包括一个根插件、两类 Handler、StopPropagation、配置回调、HTTP Handler 与静态及动态任务。
// New 登记路由与生命周期回调。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlonaBot/AlonaBot/sdk"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
	"github.com/AlonaBot/AlonaBot/sdk/page"
	"github.com/AlonaBot/AlonaBot/sdk/task"
)

// OnConfigLoad 检查 interval_seconds 非负，原样返回完整配置。
func (*Demo) OnConfigLoad(_ context.Context, _ uint32, raw json.RawMessage) (json.RawMessage, error) {
	var c struct {
		Interval int64 `json:"interval_seconds"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if c.Interval < 0 {
		return nil, errors.New("interval_seconds must not be negative")
	}
	return raw, nil
}

// Demo 作为根插件注册命令、HTTP 路由与定时任务；插件 ID 来自 plugin.yaml。
type Demo struct {
	plugin   *sdk.Plugin
	reminder task.Type[ReminderArgs]
}

// ReminderArgs 保存提醒发送所需的会话与完整文本。
type ReminderArgs struct {
	Session string `json:"session"`
	Text    string `json:"text"`
}

// New 每次创建独立的插件和示例状态。
func New() *sdk.Plugin {
	p := sdk.NewPlugin()

	demo := &Demo{plugin: p, reminder: task.Define[ReminderArgs]("remind")}
	task.On(p, demo.reminder).Handle(demo.Remind)
	p.OnConfigLoad(demo.OnConfigLoad)
	p.OnStart(demo.OnStart)
	g := p.Group("demo")
	g.Command("hello").Handle(demo.Hello)
	g.Command("echo").Handle(demo.Echo)
	g.Command("remind").Handle(demo.CreateReminder)
	g.Command("cancel").Handle(demo.CancelReminder)
	p.Route(http.MethodGet, "/status").HandleFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	p.Route(http.MethodGet, "/kit.css").Handle(page.Handler())
	p.Route(http.MethodGet, "/kit.js").Handle(page.Handler())
	p.Route(http.MethodGet, "/page/kit.css").Handle(page.Handler())
	p.Route(http.MethodGet, "/page/kit.js").Handle(page.Handler())
	p.Route(http.MethodGet, "/page").HandleFunc(demo.Page)
	p.Every("tick", time.Minute).Handle(func(*sdk.Context) error { return nil })
	return p
}

// Page 提供使用 sdk/page 套件的示例插件页面。
func (*Demo) Page(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(showcaseHTML))
}

const showcaseHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>Showcase Page</title>
  <link rel="stylesheet" href="kit.css">
  <script src="kit.js"></script>
</head>
<body>
  <div class="ab-card">
    <div class="ab-card-header">AlonaBot Showcase 插件页面</div>
    <div class="ab-card-body">
      <p>状态：<span class="ab-badge ab-badge-success" id="theme-badge">theme: light</span> <span class="ab-badge" id="locale-badge">locale: en-US</span></p>
      <p id="greeting"></p>

      <h3>控件样式示例</h3>
      <div style="display: flex; gap: 8px; margin-bottom: 16px; flex-wrap: wrap;">
        <button class="ab-btn ab-btn-primary" id="btn-request">请求 /status</button>
        <button class="ab-btn" id="btn-t">测试文案 t()</button>
        <button class="ab-btn ab-btn-danger">危险按钮</button>
        <button class="ab-btn ab-btn-soft">柔和按钮</button>
        <button class="ab-btn" disabled>禁用按钮</button>
      </div>

      <div style="margin-bottom: 16px;">
        <label for="demo-input">输入框：</label>
        <input type="text" id="demo-input" class="ab-input" placeholder="请输入内容..." value="测试输入">
      </div>

      <div style="margin-bottom: 16px;">
        <label for="demo-select">选择框：</label>
        <select id="demo-select" class="ab-select">
          <option value="1">选项一</option>
          <option value="2">选项二</option>
          <option value="3">选项三</option>
        </select>
      </div>

      <div style="margin-bottom: 16px;">
        <table class="ab-table">
          <thead>
            <tr><th>项目</th><th>状态</th><th>备注</th></tr>
          </thead>
          <tbody>
            <tr><td>路由服务</td><td><span class="ab-badge ab-badge-success">正常</span></td><td>/status 响应正常</td></tr>
            <tr><td>桥通信</td><td><span class="ab-badge ab-badge-warning" id="bridge-status">等待</span></td><td>postMessage 通道</td></tr>
            <tr><td>快捷键</td><td><span class="ab-badge ab-badge-critical">mod+k</span></td><td>捕获 ⌘K/Ctrl+K</td></tr>
          </tbody>
        </table>
      </div>

      <div class="ab-card" style="margin-bottom: 0;">
        <div class="ab-card-header">运行日志与桥反馈</div>
        <pre id="log" style="margin: 0; white-space: pre-wrap; word-break: break-all;">就绪</pre>
      </div>
    </div>
    <div class="ab-card-footer">
      提示：在页面内按 ⌘K 或 Ctrl+K 将触发快捷键转发给 WebUI。
    </div>
  </div>

  <script>
    var logEl = document.getElementById('log');
    function log(msg) {
      if (logEl) {
        logEl.textContent = new Date().toLocaleTimeString() + ' ' + msg + '\n' + logEl.textContent;
      }
    }

    async function renderGreeting() {
      try {
        var text = await window.AlonaBot.t('hello_message', { name: 'Alona' });
        document.getElementById('greeting').textContent = text;
        log('文案 t(): ' + text);
      } catch (err) {
        log('文案请求失败: ' + err.message);
      }
    }

    if (window.AlonaBot) {
      window.AlonaBot.onTheme(function(theme, tokens) {
        var badge = document.getElementById('theme-badge');
        if (badge) badge.textContent = 'theme: ' + theme;
        log('收到主题变更: ' + theme);
      });

      window.AlonaBot.onLocale(async function(locale) {
        var badge = document.getElementById('locale-badge');
        if (badge) badge.textContent = 'locale: ' + locale;
        log('收到语言变更: ' + locale);
        await renderGreeting();
      });

      var bridgeStatus = document.getElementById('bridge-status');
      if (bridgeStatus) {
        bridgeStatus.textContent = window.parent !== window ? 'iframe 桥' : '独立模式';
      }

      var btnRequest = document.getElementById('btn-request');
      if (btnRequest) {
        btnRequest.addEventListener('click', function() {
          window.AlonaBot.request('GET', '/status').then(function(res) {
            log('请求 /status 成功: status=' + res.status + ' body=' + res.body);
          }).catch(function(err) {
            log('请求 /status 失败: ' + err.message);
          });
        });
      }

      var btnT = document.getElementById('btn-t');
      if (btnT) {
        btnT.addEventListener('click', renderGreeting);
      }
    }
  </script>
</body>
</html>`

// CreateReminder 在当前用户与会话内创建提醒，回复用户输入的 key。
func (p *Demo) CreateReminder(c *sdk.Context, key string, seconds int64, text string) (message.Content, error) {
	internalKey, err := reminderKey(c, key)
	if err != nil {
		return nil, err
	}
	_, err = p.reminder.At(c.Context(), time.Now().Add(time.Duration(seconds)*time.Second),
		ReminderArgs{Session: c.Session(), Text: text}, task.Key(internalKey))
	if err != nil {
		return nil, err
	}
	return message.Text(key), nil
}

// CancelReminder 取消当前用户在当前会话的提醒，成功后回显用户输入的 key。
func (p *Demo) CancelReminder(c *sdk.Context, key string) (message.Content, error) {
	internalKey, err := reminderKey(c, key)
	if err != nil {
		return nil, err
	}
	if err := p.reminder.Cancel(c.Context(), internalKey); err != nil {
		return nil, err
	}
	return message.Text(key), nil
}

func reminderKey(c *sdk.Context, key string) (string, error) {
	user, session := c.User(), c.Session()
	// 长度前缀保留各段边界，用户 key 中的分隔符也无法改变身份作用域。
	internalKey := fmt.Sprintf("%d:%s/%d:%s/%s", len(user), user, len(session), session, key)
	if len(internalKey) > 128 { // CONTRACTS §4.3 按 UTF-8 字节限制完整内部 key。
		return "", errcode.InvalidArgument.New("reminder key with user and session scope exceeds 128 bytes")
	}
	return internalKey, nil
}

// Remind 使用参数保存的会话显式发送提醒。
func (*Demo) Remind(c *sdk.Context, args ReminderArgs) error {
	_, err := c.SendTo(args.Session, message.Text(args.Text))
	return err
}

// Hello 响应 /demo hello，返回要自动回复的 Chain；成功且非空时由 Core 发送一次。
func (p *Demo) Hello(c *sdk.Context) (message.Content, error) {
	c.StopPropagation()
	return message.Chain{message.Text(p.plugin.Config().String("greeting"))}, nil
}

// Echo 响应 /demo echo <text> <times>，在 Handler 内主动发送 times 次 text。
func (p *Demo) Echo(c *sdk.Context, text string, times int64) error {
	for i := int64(0); i < times; i++ {
		if _, err := c.Send(message.Chain{message.Text(text)}); err != nil {
			return err
		}
	}
	return nil
}

// OnStart 演示生命周期 Hook 的签名，不做任何事。
func (p *Demo) OnStart(context.Context) error { return nil }

// main 把根插件交给 SDK 运行，失败时打印原因并退出。
func main() {
	if err := sdk.Run(New()); err != nil {
		log.Fatal(err)
	}
}
