package sdk

import (
	"encoding/json"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

// AdapterClient 绑定 Handler 的可信平台身份；动作沿用父调用的权限、取消与期限。
// 使用 Context.Adapter 创建，身份在创建时冻结，调用有效期随 Handler 结束。
type AdapterClient struct {
	context    *Context
	adapterID  string
	instanceID string
	accountID  string
	session    string
}

// Adapter 取得当前可信目标或账号事件的客户端；没有平台身份时返回 NO_DEFAULT_TARGET。
func (c *Context) Adapter() (*AdapterClient, error) {
	if !c.active.Load() {
		return nil, errcode.FailedPrecondition.New("handler context has expired")
	}
	target := c.target()
	client := &AdapterClient{context: c, adapterID: target.GetAdapterId(), instanceID: target.GetAdapterInstanceId(), accountID: target.GetAccountId(), session: c.Session()}
	if target == nil && c.event != nil {
		client.adapterID = c.event.GetAdapterId()
		client.instanceID = c.event.GetAdapterInstanceId()
		client.accountID = c.event.GetAccountId()
	}
	if client.adapterID == "" || client.instanceID == "" || client.accountID == "" {
		return nil, errcode.NoDefaultTarget.New("no default adapter identity")
	}
	return client, nil
}

// AdapterID 返回绑定的平台标识。
func (a *AdapterClient) AdapterID() string { return a.adapterID }

// InstanceID 返回绑定的实例标识。
func (a *AdapterClient) InstanceID() string { return a.instanceID }

// AccountID 返回绑定的账号标识。
func (a *AdapterClient) AccountID() string { return a.accountID }

// Session 返回绑定的默认会话；账号事件为空。
func (a *AdapterClient) Session() string { return a.session }

// Call 按已声明动作范围选择当前会话或账号，保留结果与错误。
func (a *AdapterClient) Call(name string, version uint32, params any) (json.RawMessage, error) {
	return a.CallAction("", name, version, params)
}

// CallAction 实现 action.Caller；空 to 按声明选择当前会话或账号，显式 to 沿用发信许可。
// 默认使用绑定实例，Via 可显式选择其他实例并接受原有权限检查。
func (a *AdapterClient) CallAction(to string, name string, version uint32, params any, opts ...SendOption) (json.RawMessage, error) {
	options := append([]SendOption{Via(a.instanceID)}, opts...)
	if to == "" {
		options = append(options, func(o *sendOptions) { o.actionDefaultTarget = true })
	}
	return a.context.CallAction(to, name, version, params, options...)
}
