// Package debug 通过管理认证连接运行中的 Core，发现并调用 Adapter 的现有操作。
//
// Client 持有调用方提供的凭证与 HTTP 客户端；关闭请求或释放 Client 保持 Core 继续运行。
package debug

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/debugapi"
)

// Operation 描述当前运行实例的操作及 Schema。
type Operation = debugapi.Operation

// Call 指定操作、版本、目标与 JSON 参数。
type Call = debugapi.Call

// Result 保留 payload、error 与 trace_id。
type Result = debugapi.Result

// Error 保存 Core 返回的业务错误。
type Error = debugapi.Error

// Client 连接一个 Core 管理入口，可并发使用；字段在首次调用前设置。
type Client struct {
	URL   string
	Token string
	Host  string       // 反向代理前的 Core Host 校验值；空值沿用 URL。
	HTTP  *http.Client // nil 使用 http.DefaultClient。
}

// Instances 返回 Core 的实例列表，包括停止与失败的实例。
func (c *Client) Instances(ctx context.Context) (json.RawMessage, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/adapter/instances", nil)
}

// Operations 查询当前实例的真实接口；可选 name 在客户端筛选全部版本。
func (c *Client) Operations(ctx context.Context, instance, name string) ([]Operation, error) {
	raw, err := c.request(ctx, http.MethodGet, "/api/v1/debug/adapters/"+url.PathEscape(instance)+"/operations", nil)
	if err != nil {
		return nil, err
	}
	var result []Operation
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if name != "" {
		filtered := make([]Operation, 0)
		for _, op := range result {
			if op.Name == name {
				filtered = append(filtered, op)
			}
		}
		result = filtered
	}
	return result, nil
}

// Invoke 调用一次操作，部分成功时同时返回 Result 与业务 Error。
func (c *Client) Invoke(ctx context.Context, call Call) (Result, error) {
	raw, err := c.request(ctx, http.MethodPost, "/api/v1/debug/adapters/call", call)
	if err != nil {
		return Result{}, err
	}
	var result Result
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	if result.Error != nil {
		return result, result.Error
	}
	return result, nil
}

func (c *Client) request(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	base, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}
	if base.Host == "" || base.User != nil || (base.Scheme != "http" && base.Scheme != "https") || base.RawQuery != "" || base.Fragment != "" {
		return nil, fmt.Errorf("debug URL must be an absolute HTTP(S) URL without credentials, query or fragment")
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(c.URL, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	if c.Host != "" {
		req.Host = c.Host
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := http.DefaultClient
	if c.HTTP != nil {
		client = c.HTTP
	}
	local := *client
	local.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := local.Do(req)
	if err != nil {
		return nil, err
	}
	raw, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("debug API %s: %s", response.Status, raw)
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("debug API returned invalid JSON")
	}
	return json.RawMessage(raw), nil
}
