package validate

import (
	"runtime"
	"strings"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"google.golang.org/protobuf/proto"
)

func txt(s string, ents ...*pb.TextEntity) *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: s, Entities: ents}}}
}

func ent(typ pb.EntityType, off, n uint32) *pb.TextEntity {
	return &pb.TextEntity{Type: typ, OffsetUtf16: off, LengthUtf16: n}
}

func img(src *pb.MediaSource) *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: src}}}
}

func ref(id string) *pb.MediaSource {
	return &pb.MediaSource{Source: &pb.MediaSource_MediaRef{MediaRef: id}}
}

func file(name string) *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_File{File: &pb.FileSegment{Source: ref("m1"), Filename: proto.String(name)}}}
}

func reply(id string) *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_Reference{Reference: &pb.ReferenceSegment{OriginalPlatformMessageId: id}}}
}

func buttons(rows ...[]*pb.Button) *pb.Segment {
	b := &pb.ButtonsSegment{}
	for _, r := range rows {
		b.Rows = append(b.Rows, &pb.ButtonRow{Buttons: r})
	}
	return &pb.Segment{Data: &pb.Segment_Buttons{Buttons: b}}
}

func btn(action, label string, value *string) *pb.Button {
	return &pb.Button{ActionId: action, Label: label, Value: value}
}

func fwd(items ...*pb.MessageChain) *pb.Segment {
	f := &pb.ForwardSegment{}
	for _, c := range items {
		f.Items = append(f.Items, &pb.ForwardItem{Chain: c})
	}
	return &pb.Segment{Data: &pb.Segment_Forward{Forward: f}}
}

func chainOf(segs ...*pb.Segment) *pb.MessageChain { return &pb.MessageChain{Segments: segs} }

func TestChain(t *testing.T) {
	link := ent(pb.EntityType_ENTITY_TYPE_TEXT_LINK, 0, 2)
	link.Url = proto.String("https://example.com")
	mention := ent(pb.EntityType_ENTITY_TYPE_TEXT_MENTION, 0, 2)
	mention.ExternalUserId = proto.String("42")
	pre := ent(pb.EntityType_ENTITY_TYPE_PRE, 0, 2)
	pre.Language = proto.String("go")
	markdown := &pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{Text: "*hi*", Format: pb.TextFormat_TEXT_FORMAT_MARKDOWN.Enum()}}}

	cases := []struct {
		name    string
		chain   *pb.MessageChain
		wantErr string // 空串表示应通过
	}{
		// 合法
		{"纯文本", chainOf(txt("你好")), ""},
		{"markdown 文本", chainOf(markdown), ""},
		{"实体落在代理对边界", chainOf(txt("a😀b", ent(pb.EntityType_ENTITY_TYPE_BOLD, 1, 2))), ""},
		{"text_link/text_mention/pre", chainOf(txt("hi", link), txt("hi", mention), txt("hi", pre)), ""},
		{"三种媒体来源", chainOf(img(ref("m1")),
			img(&pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{Token: "in1"}}}),
			img(&pb.MediaSource{Source: &pb.MediaSource_HttpsUrl{HttpsUrl: "https://example.com/a.png"}})), ""},
		{"文件名", chainOf(file("报告.pdf")), ""},
		{"文本 + 引用 + 按钮", chainOf(reply("100"), txt("选一个"), buttons([]*pb.Button{btn("vote", "赞成", proto.String(`{"v":1}`))})), ""},
		{"合并转发", chainOf(fwd(chainOf(txt("原消息")))), ""},

		// 非法
		{"空 Chain", chainOf(), "Chain 为空"},
		{"nil Chain", nil, "Chain 为空"},
		{"片段未设置类型", chainOf(&pb.Segment{}), "没有设置类型"},
		{"只有引用和按钮", chainOf(reply("1"), buttons([]*pb.Button{btn("a", "b", nil)})), "没有可发送内容"},
		{"两个 reference", chainOf(reply("1"), reply("2"), txt("x")), "最多一个 reference"},
		{"两个 buttons", chainOf(txt("x"), buttons([]*pb.Button{btn("a", "b", nil)}), buttons([]*pb.Button{btn("a", "b", nil)})), "最多一个 buttons"},
		{"空文本", chainOf(txt("")), "text 为空"},
		{"非法 UTF-8", chainOf(txt("\xff")), "UTF-8"},
		{"format 与 entities 同时指定", chainOf(&pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{
			Text: "hi", Format: pb.TextFormat_TEXT_FORMAT_PLAIN.Enum(), Entities: []*pb.TextEntity{ent(pb.EntityType_ENTITY_TYPE_BOLD, 0, 1)}}}}), "不能同时指定"},
		{"format 显式为 UNSPECIFIED", chainOf(&pb.Segment{Data: &pb.Segment_Text{Text: &pb.TextSegment{
			Text: "hi", Format: pb.TextFormat_TEXT_FORMAT_UNSPECIFIED.Enum()}}}), "format 取值非法"},
		{"实体截断代理对", chainOf(txt("a😀b", ent(pb.EntityType_ENTITY_TYPE_BOLD, 2, 1))), "截断了字符"},
		{"实体越界", chainOf(txt("ab", ent(pb.EntityType_ENTITY_TYPE_BOLD, 1, 5))), "超出文本长度"},
		{"实体长度为 0", chainOf(txt("ab", ent(pb.EntityType_ENTITY_TYPE_BOLD, 0, 0))), "不能为 0"},
		{"实体类型 UNSPECIFIED", chainOf(txt("ab", ent(0, 0, 1))), "不支持的实体类型"},
		{"text_link 缺 url", chainOf(txt("ab", ent(pb.EntityType_ENTITY_TYPE_TEXT_LINK, 0, 1))), "url"},
		{"bold 带 url", chainOf(txt("ab", func() *pb.TextEntity {
			e := ent(pb.EntityType_ENTITY_TYPE_BOLD, 0, 1)
			e.Url = proto.String("https://x.y")
			return e
		}())), "url"},
		{"text_mention 缺用户", chainOf(txt("ab", ent(pb.EntityType_ENTITY_TYPE_TEXT_MENTION, 0, 1))), "external_user_id"},
		{"language 用于非 pre", chainOf(txt("ab", func() *pb.TextEntity {
			e := ent(pb.EntityType_ENTITY_TYPE_CODE, 0, 1)
			e.Language = proto.String("go")
			return e
		}())), "language"},
		{"媒体无来源", chainOf(img(nil)), "恰好有一个来源"},
		{"空 media_ref", chainOf(img(ref(""))), "media_ref 为空"},
		{"incoming_ref 缺 token", chainOf(img(&pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: &pb.IncomingMedia{}}})), "token"},
		{"http 媒体地址", chainOf(img(&pb.MediaSource{Source: &pb.MediaSource_HttpsUrl{HttpsUrl: "http://example.com/a.png"}})), "https"},
		{"文件名带路径", chainOf(file("../etc/passwd")), "无路径"},
		{"文件名带反斜杠", chainOf(file(`C:\a.txt`)), "无路径"},
		{"mention 缺用户", chainOf(&pb.Segment{Data: &pb.Segment_Mention{Mention: &pb.MentionSegment{}}}), "external_user_id"},
		{"reference 缺消息 ID", chainOf(reply(""), txt("x")), "original_platform_message_id"},
		{"按钮无行", chainOf(txt("x"), buttons()), "没有任何行"},
		{"按钮空行", chainOf(txt("x"), buttons([]*pb.Button{})), "为空"},
		{"按钮缺 action_id", chainOf(txt("x"), buttons([]*pb.Button{btn("", "b", nil)})), "action_id"},
		{"按钮缺 label", chainOf(txt("x"), buttons([]*pb.Button{btn("a", "", nil)})), "label"},
		{"按钮 value 非 JSON", chainOf(txt("x"), buttons([]*pb.Button{btn("a", "b", proto.String("{bad"))})), "JSON"},
		{"转发无条目", chainOf(fwd()), "没有任何条目"},
		{"转发内嵌 Chain 非法", chainOf(fwd(chainOf(txt("ok")), chainOf(txt("")))), "segment[0].forward.items[1].segment[0]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Chain(c.chain)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("应通过，实际 %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("期望包含 %q 的错误，实际 %v", c.wantErr, err)
			}
			if code(err) != pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
				t.Fatalf("错误码应为 INVALID_ARGUMENT，实际 %v", code(err))
			}
		})
	}
}

// 错误信息必须指出是哪一个 segment。
func TestChainErrorNamesSegment(t *testing.T) {
	err := Chain(chainOf(txt("ok"), txt("ok"), img(nil)))
	if err == nil || !strings.Contains(err.Error(), "segment[2]") {
		t.Fatalf("错误应定位到 segment[2]，实际 %v", err)
	}
}

// 回归：边界检查的分配量不能随文本长度放大。REVIEW-2 实测旧实现在 4 MiB ASCII 文本上分配 296 MiB，
// 在本用例的中文文本上分配约 78 MB；新实现约 2.5 KB。
func TestChainLargeTextAllocation(t *testing.T) {
	c := chainOf(txt(strings.Repeat("字", 4<<20/3), ent(pb.EntityType_ENTITY_TYPE_BOLD, 0, 1)))
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	if err := Chain(c); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&after)
	if got := after.TotalAlloc - before.TotalAlloc; got > 1<<20 {
		t.Fatalf("4 MiB 文本校验分配了 %d 字节，应与文本长度无关（上限 1 MiB）", got)
	}
}
