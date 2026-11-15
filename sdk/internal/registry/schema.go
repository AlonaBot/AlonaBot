package registry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/google/jsonschema-go/jsonschema"
)

// JSONType 推导具名 JSON 类型的协议身份与版本。
func JSONType[T any]() (string, uint32, error) {
	typ := reflect.TypeFor[T]()
	if typ.PkgPath() == "" || typ.Name() == "" || strings.Contains(typ.Name(), "[") {
		return "", 0, &errcode.Error{Code: errcode.InvalidArgument, Message: fmt.Sprintf("JSON type must be a named non-generic value type: %s", typ)}
	}
	version := uint32(1)
	if value, ok := any(*new(T)).(interface{ SchemaVersion() uint32 }); ok {
		version = value.SchemaVersion()
	}
	id := typ.PkgPath() + "." + typ.Name()
	// 选项来源的 Schema 身份由协议固定，不随 Go 包移动。
	switch id {
	case "github.com/AlonaBot/AlonaBot/sdk/options.Query":
		id = validate.OptionsQuerySchemaID
	case "github.com/AlonaBot/AlonaBot/sdk/options.Page":
		id = validate.OptionsPageSchemaID
	}
	return id, version, nil
}

// SchemaID 仅将 NewPlugin 调用包的类型对齐到 main 生成键；go test 中该包保留导入路径，导入类型的身份保持完整。
func SchemaID(schemas map[string][]byte, entryPackage, id string) string {
	if _, ok := schemas[id]; ok {
		return id
	}
	if dot := strings.LastIndexByte(id, '.'); dot >= 0 && id[:dot] == entryPackage {
		mainID := "main" + id[dot:]
		if _, ok := schemas[mainID]; ok {
			return mainID
		}
	}
	return id
}

// ResolveSchema 读取生成资源并编译校验器。
func ResolveSchema(schemas map[string][]byte, id string) (*jsonschema.Resolved, error) {
	raw, ok := schemas[id]
	if !ok {
		return nil, fmt.Errorf("service schema for %s is not generated; run alonabot build", id)
	}
	schema := &jsonschema.Schema{}
	if err := json.Unmarshal(raw, schema); err != nil {
		return nil, fmt.Errorf("decode service schema for %s: %w", id, err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		return nil, fmt.Errorf("resolve service schema for %s: %w", id, err)
	}
	return resolved, nil
}

// ValidateJSON 解码校验副本；jsonschema-go 将 json.Number 识别为 string。
func ValidateJSON(schema *jsonschema.Resolved, data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	return schema.Validate(value)
}
