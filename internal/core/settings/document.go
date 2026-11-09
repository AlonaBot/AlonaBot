package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
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
	root, err := parseNode(content, "")
	if err != nil {
		return nil, err
	}
	if root.delimiter != '{' {
		return nil, root.fail(errors.New("expected an object"))
	}
	return &Document{root: root}, nil
}

func parseNode(content []byte, path string) (*node, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	root, err := readNode(decoder, content, path)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			err = errors.New("expected a single JSON value")
		}
		return nil, jsonLocation(content, decoder.InputOffset(), path, err)
	}
	return root, nil
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
			key, _ := marshal(field.key) // 字符串总能编码为 JSON。
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
		scalar, _ := marshal(n.value) // 标量只来自 Decoder 接受的 JSON token。
		output.Write(scalar)
	}
	return output.Bytes()
}

// Set 设置路径末尾的键；缺失的父对象逐层创建，新键追加在对象末尾。
// path 中的每一项都表示对象键，数组通过设置整个值修改。
func (d *Document) Set(path []string, value any) error {
	content, err := marshal(value)
	if err != nil {
		return fmt.Errorf("%s: encode value: %w", strings.Join(path, "."), err)
	}
	replacement, err := parseNode(content, strings.Join(path, "."))
	if err != nil {
		return err
	}
	parent, err := d.parent(path, true)
	if err != nil {
		return err
	}
	key := path[len(path)-1]
	index := parent.index(key)
	if index < 0 {
		parent.members = append(parent.members, member{key: key, value: replacement})
	} else {
		parent.members[index].value = replacement
	}
	return nil
}

// Delete 删除路径末尾的键；路径缺失时保持原文，经过非对象值时返回错误。
func (d *Document) Delete(path []string) error {
	parent, err := d.parent(path, false)
	if err != nil {
		return err
	}
	if parent == nil {
		return nil
	}
	if index := parent.index(path[len(path)-1]); index >= 0 {
		parent.members = slices.Delete(parent.members, index, index+1)
	}
	return nil
}

func (d *Document) parent(path []string, create bool) (*node, error) {
	if len(path) == 0 {
		return nil, errors.New("settings: path must contain a key")
	}
	parent := d.root
	for i, key := range path[:len(path)-1] {
		if parent.delimiter != '{' {
			return nil, parent.fail(errors.New("expected an object in key path"))
		}
		index := parent.index(key)
		if index < 0 {
			if !create {
				return nil, nil
			}
			parent.members = append(parent.members, member{key: key, value: &node{delimiter: '{', path: strings.Join(path[:i+1], "."), line: parent.line, column: parent.column}})
			index = len(parent.members) - 1
		}
		parent = parent.members[index].value
	}
	if parent.delimiter != '{' {
		return nil, parent.fail(errors.New("expected an object in key path"))
	}
	return parent, nil
}

func (n *node) index(key string) int {
	return slices.IndexFunc(n.members, func(field member) bool { return field.key == key })
}

func (d *Document) formatted() ([]byte, error) {
	var content bytes.Buffer
	if err := json.Indent(&content, d.root.encode(), "", "  "); err != nil {
		return nil, err
	}
	content.WriteByte('\n')
	return content.Bytes(), nil
}

func marshal(value any) ([]byte, error) {
	var content bytes.Buffer
	encoder := json.NewEncoder(&content)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(content.Bytes(), []byte{'\n'}), nil
}
