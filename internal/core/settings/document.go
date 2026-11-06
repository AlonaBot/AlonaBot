package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

// Document 保存未展开的 JSON 文档；对象成员的顺序与文件一致。
type Document struct{ root *node }

type member struct {
	key   string
	value *node
}

type node struct {
	delimiter    json.Delim
	value        any
	members      []member
	elements     []*node
	path         string
	line, column int
}

var environmentReference = regexp.MustCompile(`^\$\{[A-Za-z_][A-Za-z0-9_]*\}$`)

func parseDocument(content []byte) (*Document, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	root, err := readNode(decoder, content, "")
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			err = errors.New("expected a single JSON object")
		}
		return nil, jsonLocation(content, decoder.InputOffset(), "", err)
	}
	if root.delimiter != '{' {
		return nil, root.fail(errors.New("expected an object"))
	}
	return &Document{root: root}, nil
}

func readNode(decoder *json.Decoder, content []byte, path string) (*node, error) {
	start := nextOffset(content, decoder.InputOffset())
	token, err := decoder.Token()
	if err != nil {
		return nil, jsonLocation(content, start, path, err)
	}
	prefix := content[:start]
	n := &node{value: token, path: path, line: bytes.Count(prefix, []byte{'\n'}) + 1, column: int(start) - bytes.LastIndexByte(prefix, '\n')}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return n, nil
	}
	n.delimiter = delimiter
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		for decoder.More() {
			keyOffset := nextOffset(content, decoder.InputOffset())
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, jsonLocation(content, keyOffset, path, err)
			}
			key := keyToken.(string) // Decoder 在对象中只接受字符串键。
			childPath := key
			if path != "" {
				childPath = path + "." + key
			}
			if seen[key] {
				return nil, jsonLocation(content, keyOffset, childPath, errors.New("duplicate key"))
			}
			seen[key] = true
			child, err := readNode(decoder, content, childPath)
			if err != nil {
				return nil, err
			}
			// 对象字段的诊断指向键，数组元素仍指向值。
			prefix := content[:keyOffset]
			child.line = bytes.Count(prefix, []byte{'\n'}) + 1
			child.column = int(keyOffset) - bytes.LastIndexByte(prefix, '\n')
			n.members = append(n.members, member{key: key, value: child})
		}
	case '[':
		for decoder.More() {
			child, err := readNode(decoder, content, path+"["+strconv.Itoa(len(n.elements))+"]")
			if err != nil {
				return nil, err
			}
			n.elements = append(n.elements, child)
		}
	default:
		return nil, n.fail(errors.New("unexpected closing delimiter"))
	}
	if _, err := decoder.Token(); err != nil {
		return nil, jsonLocation(content, decoder.InputOffset(), path, err)
	}
	return n, nil
}

// InputOffset 指向前一个 token 的结尾；跳过分隔符定位下一个键或值。
func nextOffset(content []byte, offset int64) int64 {
	for offset < int64(len(content)) {
		switch content[offset] {
		case ' ', '\n', '\r', '\t', ',', ':':
			offset++
		default:
			return offset
		}
	}
	return offset
}

func jsonLocation(content []byte, offset int64, path string, err error) error {
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		offset = syntax.Offset - 1
	}
	prefix := content[:offset]
	return (&node{path: path, line: bytes.Count(prefix, []byte{'\n'}) + 1, column: int(offset) - bytes.LastIndexByte(prefix, '\n')}).fail(err)
}

func (n *node) fail(err error) error {
	path := n.path
	if path == "" {
		path = "settings"
	}
	return fmt.Errorf("%s: line %d, column %d: %w", path, n.line, n.column, err)
}

func (n *node) object(visit func(string, *node) error) error {
	if n.delimiter != '{' {
		return n.fail(errors.New("expected an object"))
	}
	for _, field := range n.members {
		if err := visit(field.key, field.value); err != nil {
			return err
		}
	}
	return nil
}

func (n *node) text() (string, error) {
	value, ok := n.value.(string)
	if !ok {
		return "", n.fail(errors.New("expected a string"))
	}
	return value, nil
}

func (n *node) integer() (int, error) {
	number, ok := n.value.(json.Number)
	if !ok {
		return 0, n.fail(errors.New("expected an integer"))
	}
	value, err := strconv.Atoi(string(number))
	if err != nil {
		return 0, n.fail(err)
	}
	return value, nil
}

func (n *node) strings() ([]string, error) {
	if n.delimiter != '[' {
		return nil, n.fail(errors.New("expected an array of strings"))
	}
	var values []string
	for _, child := range n.elements {
		value, err := child.text()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// 展开时建立另一棵树，写回始终使用 Document 中的原始字符串。
func (n *node) expanded() (*node, error) {
	result := *n
	if value, ok := n.value.(string); ok && environmentReference.MatchString(value) {
		resolved, exists := os.LookupEnv(value[2 : len(value)-1])
		if !exists {
			return nil, n.fail(fmt.Errorf("environment variable %s is not set", value[2:len(value)-1]))
		}
		result.value = resolved
	}
	result.members = make([]member, len(n.members))
	for i, field := range n.members {
		child, err := field.value.expanded()
		if err != nil {
			return nil, err
		}
		result.members[i] = member{key: field.key, value: child}
	}
	result.elements = make([]*node, len(n.elements))
	for i, child := range n.elements {
		expanded, err := child.expanded()
		if err != nil {
			return nil, err
		}
		result.elements[i] = expanded
	}
	return &result, nil
}

func (n *node) encode() []byte {
	var output bytes.Buffer
	switch n.delimiter {
	case '{':
		output.WriteByte('{')
		for i, field := range n.members {
			if i > 0 {
				output.WriteByte(',')
			}
			key, _ := json.Marshal(field.key) // 字符串总能编码为 JSON。
			output.Write(key)
			output.WriteByte(':')
			output.Write(field.value.encode())
		}
		output.WriteByte('}')
	case '[':
		output.WriteByte('[')
		for i, child := range n.elements {
			if i > 0 {
				output.WriteByte(',')
			}
			output.Write(child.encode())
		}
		output.WriteByte(']')
	default:
		scalar, _ := json.Marshal(n.value) // 标量只来自 Decoder 接受的 JSON token。
		output.Write(scalar)
	}
	return output.Bytes()
}
