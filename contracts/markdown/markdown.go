// Package markdown 将 CommonMark 与删除线转换成普通文本或平台无关的 UTF-16 富文本实体。
package markdown

import (
	"fmt"
	"html"
	"strings"
	"unicode/utf16"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// PlainText 转换正文，保留块结构及链接、图片的说明和 URL；HTML 原文保留。
func PlainText(content string) string {
	result, _ := render(content, true)
	return result.Text
}

// RichText 转换正文及实体；图片和 HTML 返回显式错误。
func RichText(content string) (*pb.TextSegment, error) { return render(content, false) }

type renderer struct {
	source   []byte
	plain    bool
	prefix   string
	out      strings.Builder
	units    uint32
	entities []*pb.TextEntity
}

func (r *renderer) write(s string) {
	for len(s) > 0 {
		i := strings.IndexByte(s, '\n')
		chunk := s
		if i >= 0 {
			chunk = s[:i+1]
		}
		if chunk != "\n" && (r.out.Len() == 0 || strings.HasSuffix(r.out.String(), "\n")) {
			r.out.WriteString(r.prefix)
			r.units += uint32(len(utf16.Encode([]rune(r.prefix))))
		}
		r.out.WriteString(chunk)
		r.units += uint32(len(utf16.Encode([]rune(chunk))))
		s = s[len(chunk):]
	}
}
func (r *renderer) newline() {
	if r.out.Len() > 0 && !strings.HasSuffix(r.out.String(), "\n") {
		r.write("\n")
	}
}
func (r *renderer) children(n ast.Node) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := r.node(c); err != nil {
			return err
		}
	}
	return nil
}
func (r *renderer) node(n ast.Node) error {
	start := r.units
	var entity *pb.TextEntity
	switch v := n.(type) {
	case *ast.Text:
		s := v.Segment.Value(r.source)
		if !v.IsRaw() {
			s = util.UnescapePunctuations(s)
			r.write(html.UnescapeString(string(s)))
		} else {
			r.write(string(s))
		}
		if v.SoftLineBreak() || v.HardLineBreak() {
			r.write("\n")
		}
		return nil
	case *ast.String:
		r.write(string(v.Value))
		return nil
	case *ast.Image:
		if !r.plain {
			return fmt.Errorf("markdown image cannot be represented by text entities")
		}
		if err := r.children(n); err != nil {
			return err
		}
		r.write(" (" + string(v.Destination) + ")")
		return nil
	case *ast.RawHTML:
		if !r.plain {
			return fmt.Errorf("markdown HTML cannot be represented by text entities")
		}
		for i := 0; i < v.Segments.Len(); i++ {
			s := v.Segments.At(i)
			r.write(string(s.Value(r.source)))
		}
		return nil
	case *ast.HTMLBlock:
		if !r.plain {
			return fmt.Errorf("markdown HTML cannot be represented by text entities")
		}
		for i := 0; i < v.Lines().Len(); i++ {
			s := v.Lines().At(i)
			r.write(string(s.Value(r.source)))
		}
		if v.HasClosure() {
			r.write(string(v.ClosureLine.Value(r.source)))
		}
		return nil
	case *ast.Emphasis:
		typ := pb.EntityType_ENTITY_TYPE_ITALIC
		if v.Level == 2 {
			typ = pb.EntityType_ENTITY_TYPE_BOLD
		}
		entity = &pb.TextEntity{Type: typ}
	case *extast.Strikethrough:
		entity = &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_STRIKETHROUGH}
	case *ast.CodeSpan:
		for child := v.FirstChild(); child != nil; child = child.NextSibling() {
			t := child.(*ast.Text)
			r.write(strings.ReplaceAll(string(t.Segment.Value(r.source)), "\n", " "))
		}
		if !r.plain && r.units > start {
			r.entities = append(r.entities, &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_CODE, OffsetUtf16: start, LengthUtf16: r.units - start})
		}
		return nil
	case *ast.Link:
		entity = &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_TEXT_LINK, Url: protoString(string(v.Destination))}
	case *ast.AutoLink:
		r.write(string(v.Label(r.source)))
		if r.plain {
			r.write(" (" + string(v.URL(r.source)) + ")")
		} else {
			r.entities = append(r.entities, &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_TEXT_LINK, OffsetUtf16: start, LengthUtf16: r.units - start, Url: protoString(string(v.URL(r.source)))})
		}
		return nil
	case *ast.FencedCodeBlock:
		r.newline()
		start = r.units
		for i := 0; i < v.Lines().Len(); i++ {
			s := v.Lines().At(i)
			r.write(string(s.Value(r.source)))
		}
		if !r.plain && r.units > start {
			e := &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_PRE, OffsetUtf16: start, LengthUtf16: r.units - start}
			if lang := v.Language(r.source); len(lang) > 0 {
				e.Language = protoString(string(lang))
			}
			r.entities = append(r.entities, e)
		}
		r.newline()
		return nil
	case *ast.CodeBlock:
		r.newline()
		start = r.units
		for i := 0; i < v.Lines().Len(); i++ {
			s := v.Lines().At(i)
			r.write(string(s.Value(r.source)))
		}
		if !r.plain && r.units > start {
			r.entities = append(r.entities, &pb.TextEntity{Type: pb.EntityType_ENTITY_TYPE_PRE, OffsetUtf16: start, LengthUtf16: r.units - start})
		}
		r.newline()
		return nil
	case *ast.Heading:
		r.newline()
		r.write(strings.Repeat("#", v.Level) + " ")
	case *ast.Blockquote:
		r.newline()
		old := r.prefix
		r.prefix += "> "
		err := r.children(n)
		r.prefix = old
		r.newline()
		return err
	case *ast.ListItem:
		r.newline()
		list := v.Parent().(*ast.List)
		marker := "- "
		if list.IsOrdered() {
			index := list.Start
			for c := v.PreviousSibling(); c != nil; c = c.PreviousSibling() {
				index++
			}
			marker = fmt.Sprintf("%d. ", index)
		}
		r.write(marker)
		old := r.prefix
		r.prefix += strings.Repeat(" ", len(marker))
		err := r.children(n)
		r.prefix = old
		r.newline()
		return err
	case *ast.ThematicBreak:
		r.newline()
		r.write("---\n")
		return nil
	}
	if err := r.children(n); err != nil {
		return err
	}
	if entity != nil {
		if r.plain {
			if link, ok := n.(*ast.Link); ok {
				r.write(" (" + string(link.Destination) + ")")
			}
		} else if r.units > start {
			entity.OffsetUtf16 = start
			entity.LengthUtf16 = r.units - start
			r.entities = append(r.entities, entity)
		}
	}
	switch n.(type) {
	case *ast.Paragraph, *ast.Heading:
		r.newline()
		if _, ok := n.Parent().(*ast.ListItem); !ok {
			r.write("\n")
		}
	case *ast.List, *ast.Blockquote:
		r.newline()
	}
	return nil
}
func protoString(s string) *string { return &s }
func render(content string, plain bool) (*pb.TextSegment, error) {
	source := []byte(content)
	document := goldmark.New(goldmark.WithExtensions(extension.Strikethrough)).Parser().Parse(text.NewReader(source))
	r := &renderer{source: source, plain: plain}
	err := r.children(document)
	// 尾部块分隔保留在实体之外；实体中的代码换行不截断。
	value := r.out.String()
	var lastEntity uint32
	for _, entity := range r.entities {
		if end := entity.OffsetUtf16 + entity.LengthUtf16; end > lastEntity {
			lastEntity = end
		}
	}
	for strings.HasSuffix(value, "\n") && r.units > lastEntity {
		value = strings.TrimSuffix(value, "\n")
		r.units--
	}

	return &pb.TextSegment{Text: value, Entities: r.entities}, err
}
