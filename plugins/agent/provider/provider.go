package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type Protocol string
const (
	OpenAICompatible Protocol = "openai-compatible"
	AnthropicMessages Protocol = "anthropic-messages"
)
type Config struct { Protocol Protocol; BaseURL, APIKey, Model string }
type Message struct { Role, Content string; ToolCalls []ToolCall; ToolResults []ToolResult }
type Tool struct { Name, Description string; InputSchema json.RawMessage }
type Request struct { Messages []Message; Tools []Tool; MaxTokens int }
type ToolCall struct { ID, Name string; Arguments json.RawMessage }
type ToolResult struct { ToolCallID, Content string }
type EventKind uint8
const ( TextDelta EventKind = iota+1; ToolCallComplete )
type Event struct { Kind EventKind; Text string; ToolCall *ToolCall }
type Client interface { Stream(context.Context, Request, func(Event) error) error }
type client struct { cfg Config; http *http.Client; base *url.URL }

func New(cfg Config, hc *http.Client) (Client,error) {
	if cfg.Protocol!=OpenAICompatible && cfg.Protocol!=AnthropicMessages { return nil, errors.New("unsupported provider protocol") }
	if cfg.Model=="" || cfg.APIKey=="" { return nil, errors.New("provider model and API key are required") }
	if hc==nil { return nil, errors.New("HTTP client is required") }
	u,err:=url.Parse(cfg.BaseURL)
	if err!=nil || u.Host=="" || (u.Scheme!="http" && u.Scheme!="https") || u.User!=nil || u.RawQuery!="" || u.Fragment!="" { return nil,errors.New("provider base URL must be HTTP(S) without credentials, query, or fragment") }
	return &client{cfg:cfg,http:hc,base:u},nil
}
func (c *client) Stream(ctx context.Context, req Request, emit func(Event)error) error {
	if emit==nil { return errors.New("event callback is required") }
	if req.MaxTokens<=0 || len(req.Messages)==0 { return errors.New("messages and positive max tokens are required") }
	for _,m:=range req.Messages {
		if m.Role!="system" && m.Role!="user" && m.Role!="assistant" { return errors.New("invalid message role") }
		if len(m.ToolCalls)>0 && m.Role!="assistant" { return errors.New("tool calls require assistant role") }
		if len(m.ToolResults)>0 && m.Role!="user" { return errors.New("tool results require user role") }
		for _,r:=range m.ToolResults { if r.ToolCallID=="" { return errors.New("tool result call ID is required") } }
	}
	names:=map[string]bool{}
	for _,t:=range req.Tools {
		var schema map[string]json.RawMessage
		if t.Name=="" || names[t.Name] { return errors.New("tool names must be non-empty and unique") }; names[t.Name]=true
		if len(t.InputSchema)==0 || json.Unmarshal(t.InputSchema,&schema)!=nil || schema==nil { return errors.New("tool input schema must be a JSON object") }
	}
	if err:=ctx.Err(); err!=nil { return err }
	endpoint:="messages"; var body any
	if c.cfg.Protocol==OpenAICompatible { endpoint="chat/completions"; body=c.openAI(req) } else { body=c.anthropic(req) }
	raw,err:=json.Marshal(body); if err!=nil { return fmt.Errorf("encode provider request: %w",err) }
	u:=*c.base; u.Path=strings.TrimRight(u.Path,"/")+"/"+endpoint; u.RawPath=""
	hreq,err:=http.NewRequestWithContext(ctx,http.MethodPost,u.String(),bytes.NewReader(raw)); if err!=nil { return errors.New("build provider request") }
	hreq.Header.Set("Content-Type","application/json"); hreq.Header.Set("Accept","text/event-stream")
	if c.cfg.Protocol==OpenAICompatible { hreq.Header.Set("Authorization","Bearer "+c.cfg.APIKey) } else { hreq.Header.Set("x-api-key",c.cfg.APIKey); hreq.Header.Set("anthropic-version","2023-06-01") }
	resp,err:=c.http.Do(hreq); if err!=nil { if ctx.Err()!=nil{return ctx.Err()}; return errors.New("provider request failed") }; defer resp.Body.Close()
	if resp.StatusCode<200 || resp.StatusCode>=300 { return fmt.Errorf("provider returned HTTP status %d",resp.StatusCode) }
	if c.cfg.Protocol==OpenAICompatible { return streamOpenAI(resp.Body,emit) }; return streamAnthropic(resp.Body,emit)
}
func (c *client) openAI(r Request) map[string]any {
	msgs:=[]map[string]any{}
	for _,m:=range r.Messages {
		item:=map[string]any{"role":m.Role,"content":m.Content}
		if len(m.ToolCalls)>0 { calls:=[]map[string]any{}; for _,tc:=range m.ToolCalls { calls=append(calls,map[string]any{"id":tc.ID,"type":"function","function":map[string]any{"name":tc.Name,"arguments":string(tc.Arguments)}}) }; item["tool_calls"]=calls }
		if len(m.ToolResults)>0 { for _,tr:=range m.ToolResults { msgs=append(msgs,map[string]any{"role":"tool","tool_call_id":tr.ToolCallID,"content":tr.Content}) } } else { msgs=append(msgs,item) }
	}
	out:=map[string]any{"model":c.cfg.Model,"messages":msgs,"max_tokens":r.MaxTokens,"stream":true}
	if len(r.Tools)>0 { ts:=[]map[string]any{}; for _,t:=range r.Tools { ts=append(ts,map[string]any{"type":"function","function":map[string]any{"name":t.Name,"description":t.Description,"parameters":json.RawMessage(t.InputSchema)}}) }; out["tools"]=ts }
	return out
}
func (c *client) anthropic(r Request) map[string]any {
	msgs:=[]map[string]any{}; systems:=[]string{}
	for _,m:=range r.Messages {
		if m.Role=="system" { systems=append(systems,m.Content); continue }
		item:=map[string]any{"role":m.Role,"content":m.Content}
		if len(m.ToolCalls)>0 { bs:=[]map[string]any{}; for _,tc:=range m.ToolCalls { var input any; _=json.Unmarshal(tc.Arguments,&input); bs=append(bs,map[string]any{"type":"tool_use","id":tc.ID,"name":tc.Name,"input":input}) }; item["content"]=bs }
		if len(m.ToolResults)>0 { bs:=[]map[string]any{}; for _,tr:=range m.ToolResults { bs=append(bs,map[string]any{"type":"tool_result","tool_use_id":tr.ToolCallID,"content":tr.Content}) }; item["content"]=bs }
		msgs=append(msgs,item)
	}
	out:=map[string]any{"model":c.cfg.Model,"messages":msgs,"max_tokens":r.MaxTokens,"stream":true}
	if len(systems)>0 { out["system"]=strings.Join(systems,"\n\n") }
	if len(r.Tools)>0 { ts:=[]map[string]any{}; for _,t:=range r.Tools { ts=append(ts,map[string]any{"name":t.Name,"description":t.Description,"input_schema":json.RawMessage(t.InputSchema)}) }; out["tools"]=ts }
	return out
}
func readSSE(r io.Reader, handle func(string,string)error) error {
	br:=bufio.NewReader(r); event:=""; var data strings.Builder
	for { line,err:=br.ReadString('\n'); if len(line)>0 { line=strings.TrimSuffix(strings.TrimSuffix(line,"\n"),"\r")
		if line=="" { if data.Len()>0 { if e:=handle(event,strings.TrimSuffix(data.String(),"\n"));e!=nil{return e} }; event=""; data.Reset()
		} else if strings.HasPrefix(line,":") { } else if k,v,ok:=strings.Cut(line,":");ok { v=strings.TrimPrefix(v," "); switch k { case "event":event=v; case "data":data.WriteString(v);data.WriteByte('\n') } }
	}; if err!=nil { if errors.Is(err,io.EOF){return errors.New("provider stream ended before terminal event")};return errors.New("read provider stream") } }
}
type oaChunk struct { Choices []struct { Delta struct { Content *string `json:"content"`; ToolCalls []struct { Index int `json:"index"`; ID string `json:"id"`; Function struct { Name string `json:"name"`; Arguments string `json:"arguments"` } `json:"function"` } `json:"tool_calls"` } `json:"delta"` } `json:"choices"`; Error json.RawMessage `json:"error"` }
func streamOpenAI(r io.Reader,emit func(Event)error) error {
	calls:=map[int]*ToolCall{}; terminal:=false
	err:=readSSE(r,func(_,data string)error{
		if data=="[DONE]" { terminal=true;return nil }
		var chunk oaChunk; if json.Unmarshal([]byte(data),&chunk)!=nil || len(chunk.Error)>0 { return errors.New("invalid or failed OpenAI-compatible stream event") }
		for _,ch:=range chunk.Choices { if ch.Delta.Content!=nil && *ch.Delta.Content!="" { if err:=emit(Event{Kind:TextDelta,Text:*ch.Delta.Content});err!=nil{return err} }
			for _,d:=range ch.Delta.ToolCalls { tc:=calls[d.Index];if tc==nil{tc=&ToolCall{};calls[d.Index]=tc};tc.ID+=d.ID;tc.Name+=d.Function.Name;tc.Arguments=append(tc.Arguments,[]byte(d.Function.Arguments)...)}
		};return nil
	}); if err!=nil{return err};if !terminal{return errors.New("OpenAI-compatible stream is missing [DONE]")}
	indexes:=make([]int,0,len(calls));for i:=range calls{indexes=append(indexes,i)};sort.Ints(indexes)
	for _,i:=range indexes { tc:=calls[i];var obj map[string]json.RawMessage;if tc.ID==""||tc.Name==""||json.Unmarshal(tc.Arguments,&obj)!=nil||obj==nil{return errors.New("invalid tool call in provider stream")};copyTC:=*tc;copyTC.Arguments=append(json.RawMessage(nil),tc.Arguments...);if err:=emit(Event{Kind:ToolCallComplete,ToolCall:&copyTC});err!=nil{return err} }
	return nil
}
func streamAnthropic(r io.Reader,emit func(Event)error) error {
	type block struct{id,name string;args strings.Builder}
	blocks:=map[int]*block{};terminal:=false
	err:=readSSE(r,func(event,data string)error{
		var p struct { Type string `json:"type"`; Index int `json:"index"`; ContentBlock struct{Type,ID,Name string `json:"type"`} `json:"content_block"`; Delta struct{Type string `json:"type"`;Text string `json:"text"`;PartialJSON string `json:"partial_json"`} `json:"delta"`; Error json.RawMessage `json:"error"` }
		if json.Unmarshal([]byte(data),&p)!=nil{return errors.New("invalid Anthropic stream event")};if event=="error"||p.Type=="error"||len(p.Error)>0{return errors.New("Anthropic provider stream reported an error")}
		switch p.Type {
		case "content_block_start": if p.ContentBlock.Type=="tool_use"{blocks[p.Index]=&block{id:p.ContentBlock.ID,name:p.ContentBlock.Name}}
		case "content_block_delta": if p.Delta.Type=="text_delta"&&p.Delta.Text!=""{return emit(Event{Kind:TextDelta,Text:p.Delta.Text})};if p.Delta.Type=="input_json_delta"{b:=blocks[p.Index];if b==nil{return errors.New("Anthropic tool delta has no matching block")};b.args.WriteString(p.Delta.PartialJSON)}
		case "content_block_stop": if b:=blocks[p.Index];b!=nil{raw:=[]byte(b.args.String());var obj map[string]json.RawMessage;if b.id==""||b.name==""||json.Unmarshal(raw,&obj)!=nil||obj==nil{return errors.New("invalid tool call in Anthropic stream")};tc:=&ToolCall{ID:b.id,Name:b.name,Arguments:append(json.RawMessage(nil),raw...)};if err:=emit(Event{Kind:ToolCallComplete,ToolCall:tc});err!=nil{return err};delete(blocks,p.Index)}
		case "message_stop":terminal=true
		case "message_start","message_delta","ping","": 
		default:return fmt.Errorf("unsupported Anthropic stream event %q",p.Type)
		};return nil
	});if err!=nil{return err};if !terminal{return errors.New("Anthropic stream is missing message_stop")};if len(blocks)>0{return errors.New("Anthropic stream has unfinished content blocks")};return nil
}
