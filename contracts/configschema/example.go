package configschema

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
)

// Example 保存输入文字、按顺序排列的回复片段与全部相等时显示的条件。
type Example struct {
	Input     string
	Reply     []ExampleSegment
	Condition map[string]json.RawMessage
}

// ExampleSegment 保存一段文字或一张包内图片；Text 与 Image 恰有一个。
type ExampleSegment struct {
	Text  string
	Image string
	Alt   string
}

// ExampleImages 返回所有示例引用的 examples/ 相对路径，去重并排序。
func (s *Schema) ExampleImages() []string {
	var images []string
	var collect func([]*Field)
	collect = func(fields []*Field) {
		for _, f := range fields {
			for _, example := range f.Examples {
				for _, segment := range example.Reply {
					if segment.Image != "" {
						images = append(images, segment.Image)
					}
				}
			}
			collect(f.items)
			for _, template := range f.templates {
				collect(template.items)
			}
		}
	}
	collect(s.fields)
	slices.Sort(images)
	return slices.Compact(images)
}

func (s *Schema) checkExamples(fields []*Field, parent string, problems *[]Problem) {
	for _, f := range fields {
		fieldPath := joinPath(parent, f.name)
		if f.attributes.has("example") {
			examplePath := joinPath(fieldPath, "example")
			if examples, ok := f.attributes.get("example").([]any); ok {
				for i, value := range examples {
					f.Examples = append(f.Examples, s.parseExample(value, fmt.Sprintf("%s[%d]", examplePath, i), problems))
				}
			} else {
				f.Examples = []Example{s.parseExample(f.attributes.get("example"), examplePath, problems)}
			}
		}
		s.checkExamples(f.items, joinPath(fieldPath, "items"), problems)
		for _, template := range f.templates {
			s.checkExamples(template.items, joinPath(joinPath(joinPath(fieldPath, "templates"), template.name), "items"), problems)
		}
	}
}

func (s *Schema) parseExample(value any, location string, problems *[]Problem) Example {
	// 1. 检查示例对象的属性与必填输入文案。
	var example Example
	object, ok := value.(orderedObject)
	if !ok {
		*problems = append(*problems, Problem{location, "expected an example object or array of example objects"})
		return example
	}
	for _, attribute := range object {
		if !slices.Contains([]string{"input", "reply", "condition"}, attribute.name) {
			*problems = append(*problems, Problem{joinPath(location, attribute.name), "unknown example attribute"})
		}
	}
	example.Input = s.exampleText(object.get("input"), joinPath(location, "input"), problems)
	// 2. 将两种回复写法归一为片段，同时定位图片与占位符错误。
	replyPath := joinPath(location, "reply")
	if text, ok := object.get("reply").(string); ok {
		example.Reply = []ExampleSegment{{Text: s.exampleText(text, replyPath, problems)}}
	} else if segments, ok := object.get("reply").([]any); ok {
		for i, value := range segments {
			segmentPath := fmt.Sprintf("%s[%d]", replyPath, i)
			segment, ok := value.(orderedObject)
			if !ok {
				*problems = append(*problems, Problem{segmentPath, "expected a reply segment object"})
				continue
			}
			if segment.has("text") == segment.has("image") {
				*problems = append(*problems, Problem{segmentPath, "expected exactly one of text and image"})
			}
			var parsed ExampleSegment
			for _, attribute := range segment {
				attributePath := joinPath(segmentPath, attribute.name)
				switch attribute.name {
				case "text":
					parsed.Text = s.exampleText(attribute.value, attributePath, problems)
				case "image":
					parsed.Image, ok = attribute.value.(string)
					if !ok || !validExampleImage(parsed.Image) {
						*problems = append(*problems, Problem{attributePath, "expected a canonical examples/ path to a PNG, JPEG, WebP or GIF"})
					}
				case "alt":
					parsed.Alt, ok = attribute.value.(string)
					if !ok || !segment.has("image") {
						*problems = append(*problems, Problem{attributePath, "alt must be a string on an image segment"})
					}
				default:
					*problems = append(*problems, Problem{attributePath, "unknown reply segment attribute"})
				}
			}
			example.Reply = append(example.Reply, parsed)
		}
	} else {
		*problems = append(*problems, Problem{replyPath, "expected a reply string or segment array"})
	}
	// 3. 条件按全局字段路径解析，期望值只复用字段的类型约束。
	if object.has("condition") {
		conditionPath := joinPath(location, "condition")
		condition, ok := object.get("condition").(orderedObject)
		if !ok {
			*problems = append(*problems, Problem{conditionPath, "expected an example condition object"})
			return example
		}
		example.Condition = make(map[string]json.RawMessage, len(condition))
		for _, expected := range condition {
			location := joinPath(conditionPath, expected.name)
			f := s.exampleField(expected.name)
			if f == nil {
				*problems = append(*problems, Problem{location, "example condition must reference a non-secret scalar field or template.$key"})
				continue
			}
			if problem := valueProblem(conditionField(f), plainValue(expected.value), location); problem != nil {
				*problems = append(*problems, *problem)
			}
			example.Condition[expected.name] = appendCanonical(nil, expected.value)
		}
	}
	return example
}

func validExampleImage(name string) bool {
	if !strings.HasPrefix(name, "examples/") || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:\x00") {
		return false
	}
	return slices.Contains([]string{".png", ".jpg", ".jpeg", ".webp", ".gif"}, path.Ext(name))
}

func (s *Schema) exampleField(reference string) *Field {
	parts := strings.Split(reference, ".")
	fields := s.fields
	for i := 0; i < len(parts); i++ {
		f := findField(fields, parts[i])
		if f == nil {
			return nil
		}
		if secret, _ := f.attributes.get("secret").(bool); secret {
			return nil
		}
		if i == len(parts)-1 {
			if slices.Contains([]string{typeString, typeText, typeInt, typeFloat, typeBool}, f.typ) {
				return f
			}
			return nil
		}
		switch f.typ {
		case typeObject:
			fields = f.items
		case typeTemplate:
			i++
			if parts[i] == "$key" && i == len(parts)-1 {
				return &Field{typ: typeString}
			}
			fields = nil
			for _, template := range f.templates {
				if template.name == parts[i] {
					fields = template.items
					break
				}
			}
		default:
			return nil
		}
	}
	return nil
}

func (s *Schema) exampleText(value any, location string, problems *[]Problem) string {
	text, ok := value.(string)
	if !ok {
		*problems = append(*problems, Problem{location, "expected example text"})
		return ""
	}
	for i := 0; i < len(text); {
		if text[i] != '{' && text[i] != '}' {
			i++
			continue
		}
		if i+1 < len(text) && text[i] == text[i+1] {
			i += 2
			continue
		}
		end := strings.IndexByte(text[i+1:], '}')
		if text[i] == '}' || end < 0 {
			*problems = append(*problems, Problem{location, "unescaped or unmatched example brace"})
			break
		}
		end += i + 1
		if s.exampleField(text[i+1:end]) == nil {
			*problems = append(*problems, Problem{location, "example placeholder must reference a non-secret scalar field or template.$key"})
		}
		i = end + 1
	}
	return text
}
