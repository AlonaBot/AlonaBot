package message

import "strings"

// Session 保存平台会话及可选话题，可直接用字符串字面量声明。
type Session string

var sessionEscape = strings.NewReplacer("%", "%25", "#", "%23")

// NewSession 转义聊天和话题中的保留字符，threadID 为空时省略话题。
func NewSession(adapterID, chatID, threadID string) Session {
	s := adapterID + ":" + sessionEscape.Replace(chatID)
	if threadID != "" {
		s += "#" + sessionEscape.Replace(threadID)
	}
	return Session(s)
}

// Parts 拆出 Adapter、聊天与话题；格式不合法时 ok 为 false。
func (s Session) Parts() (adapterID, chatID, threadID string, ok bool) {
	adapterID, rest, found := strings.Cut(string(s), ":")
	if !found || adapterID == "" || rest == "" || strings.ContainsAny(adapterID, "#%/@") {
		return "", "", "", false
	}
	chat, thread, topic := strings.Cut(rest, "#")
	if chat == "" || (topic && thread == "") || strings.Contains(thread, "#") {
		return "", "", "", false
	}
	chatID, ok = unescape(chat, map[string]string{"25": "%", "23": "#"})
	if !ok {
		return "", "", "", false
	}
	threadID, ok = unescape(thread, map[string]string{"25": "%", "23": "#"})
	return adapterID, chatID, threadID, ok
}

// MessageRef 保存一条消息的会话、平台消息 ID 与可选的机器人实例。
type MessageRef string

var messageEscape = strings.NewReplacer("%", "%25", "/", "%2F", "@", "%40")

// NewMessageRef 构造可存储的消息句柄，省略空实例。
func NewMessageRef(s Session, platformMessageID, instanceID string) MessageRef {
	r := string(s) + "/" + messageEscape.Replace(platformMessageID)
	if instanceID != "" {
		r += "@" + instanceID
	}
	return MessageRef(r)
}

// Parts 拆出会话、平台消息 ID 与实例；同时校验内嵌会话。
func (r MessageRef) Parts() (s Session, platformMessageID, instanceID string, ok bool) {
	separator := strings.LastIndexByte(string(r), '/')
	if separator < 0 {
		return "", "", "", false
	}
	s = Session(r[:separator])
	if _, _, _, valid := s.Parts(); !valid {
		return "", "", "", false
	}
	id, instance, explicit := strings.Cut(string(r[separator+1:]), "@")
	if id == "" || (explicit && !validInstance(instance)) {
		return "", "", "", false
	}
	platformMessageID, ok = unescape(id, map[string]string{"25": "%", "2F": "/", "40": "@"})
	return s, platformMessageID, instance, ok
}

func unescape(s string, escapes map[string]string) (string, bool) {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			out.WriteByte(s[i])
			continue
		}
		if i+2 >= len(s) {
			return "", false
		}
		decoded, ok := escapes[s[i+1:i+3]]
		if !ok {
			return "", false
		}
		out.WriteString(decoded)
		i += 2
	}
	return out.String(), true
}

func validInstance(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// User 保存用户在某个平台上的字符串身份。
type User string

// NewUser 保留外部用户 ID 的原文，包括其中的冒号。
func NewUser(adapterID, externalUserID string) User {
	return User(adapterID + ":" + externalUserID)
}

// Parts 按第一个冒号拆分身份，任一部分为空时 ok 为 false。
func (u User) Parts() (adapterID, externalUserID string, ok bool) {
	adapterID, externalUserID, ok = strings.Cut(string(u), ":")
	return adapterID, externalUserID, ok && adapterID != "" && externalUserID != ""
}
