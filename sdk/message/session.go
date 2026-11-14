package message

import (
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
)

var sessionEscape = strings.NewReplacer("%", "%25", "#", "%23")

// SID 拼接会话字符串，转义聊天和话题中的保留字符；空话题省略。
func SID(adapterID, chatID, threadID string) string {
	s := adapterID + ":" + sessionEscape.Replace(chatID)
	if threadID != "" {
		s += "#" + sessionEscape.Replace(threadID)
	}
	return s
}

// ParseSID 拆出 Adapter、聊天与话题；格式不合法时返回 INVALID_ARGUMENT。
func ParseSID(sid string) (adapterID, chatID, threadID string, err error) {
	invalid := errcode.InvalidArgument.New("invalid session string")
	adapterID, rest, found := strings.Cut(sid, ":")
	if !found || !validAdapterID(adapterID) || rest == "" {
		return "", "", "", invalid
	}
	chat, thread, topic := strings.Cut(rest, "#")
	if chat == "" || (topic && thread == "") || strings.Contains(thread, "#") {
		return "", "", "", invalid
	}
	chatID, ok := unescape(chat, map[string]string{"25": "%", "23": "#"})
	if !ok {
		return "", "", "", invalid
	}
	threadID, ok = unescape(thread, map[string]string{"25": "%", "23": "#"})
	if !ok {
		return "", "", "", invalid
	}
	return adapterID, chatID, threadID, nil
}

func validAdapterID(id string) bool {
	// Adapter ID 只采用本地规则；校验副本替换等长分隔符，跳过 PluginID 的市场前缀专用规则。
	if strings.HasPrefix(id, validate.MarketPrefix) {
		id = strings.Replace(id, "_", "-", 1)
	}
	_, err := validate.PluginID(id)
	return err == nil
}

var messageEscape = strings.NewReplacer("%", "%25", "/", "%2F", "@", "%40")

// Ref 拼接消息句柄，转义平台消息 ID 中的保留字符；空实例省略。
func Ref(sid, platformMessageID, instance string) string {
	r := sid + "/" + messageEscape.Replace(platformMessageID)
	if instance != "" {
		r += "@" + instance
	}
	return r
}

// ParseRef 拆出会话、平台消息 ID 与实例，并校验内嵌会话；格式不合法时返回 INVALID_ARGUMENT。
func ParseRef(ref string) (sid, platformMessageID, instance string, err error) {
	invalid := errcode.InvalidArgument.New("invalid message reference string")
	separator := strings.LastIndexByte(ref, '/')
	if separator < 0 {
		return "", "", "", invalid
	}
	sid = ref[:separator]
	if _, _, _, err := ParseSID(sid); err != nil {
		return "", "", "", invalid
	}
	id, instance, explicit := strings.Cut(ref[separator+1:], "@")
	if id == "" || (explicit && !validInstance(instance)) {
		return "", "", "", invalid
	}
	platformMessageID, ok := unescape(id, map[string]string{"25": "%", "2F": "/", "40": "@"})
	if !ok {
		return "", "", "", invalid
	}
	return sid, platformMessageID, instance, nil
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

// UID 拼接用户字符串，保留外部用户 ID 的原文，包括其中的冒号。
func UID(adapterID, externalUserID string) string { return adapterID + ":" + externalUserID }

// ParseUID 按第一个冒号拆分身份；Adapter ID 不合法或外部用户 ID 为空时返回 INVALID_ARGUMENT。
func ParseUID(uid string) (adapterID, externalUserID string, err error) {
	adapterID, externalUserID, ok := strings.Cut(uid, ":")
	if !ok || !validAdapterID(adapterID) || externalUserID == "" {
		return "", "", errcode.InvalidArgument.New("invalid user string")
	}
	return adapterID, externalUserID, nil
}
