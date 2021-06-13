package json_diff

import (
	"encoding/json"
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
	"github.com/pkg/errors"
)

func parse(v interface{}, level int64) *decode.JsonNode {
	var root *decode.JsonNode
	switch v.(type) {
	case map[string]interface{}:
		value := v.(map[string]interface{})
		root = &decode.JsonNode{Type: decode.JsonNodeTypeObject, Level: level, ChildrenMap: make(map[string]*decode.JsonNode)}
		for key, va := range value {
			key = decode.KeyReplace(key)
			n := parse(va, level+1)
			n.Key = key
			// n.Father = root
			root.ChildrenMap[key] = n
			// root.Children = append(root.Children, n)
		}
	case []interface{}:
		root = &decode.JsonNode{Type: decode.JsonNodeTypeSlice, Level: level}
		value := v.([]interface{})
		for _, va := range value {
			n := parse(va, level+1)
			// n.Father = root
			root.Children = append(root.Children, n)
		}
	default:
		root = &decode.JsonNode{Type: decode.JsonNodeTypeValue, Level: level}
		root.Type = decode.JsonNodeTypeValue
		root.Value = v
	}
	return root
}

// Parse 于 Unmarshal 无异
func Parse(input []byte) (*decode.JsonNode, error) {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}
	return parse(v, 0), nil
}

func marshalValue(root *decode.JsonNode) interface{} {
	return root.Value
}

func marshalObject(root *decode.JsonNode) map[string]interface{} {
	dict := make(map[string]interface{})
	for k, v := range root.ChildrenMap {
		switch v.Type {
		case decode.JsonNodeTypeObject:
			dict[k] = marshalObject(v)
		case decode.JsonNodeTypeSlice:
			dict[k] = marshalSlice(v)
		case decode.JsonNodeTypeValue:
			dict[k] = marshalValue(v)
		}
	}
	return dict
}

func marshalSlice(root *decode.JsonNode) []interface{} {
	res := make([]interface{}, len(root.Children))
	for i, child := range root.Children {
		switch child.Type {
		case decode.JsonNodeTypeValue:
			res[i] = marshalValue(child)
		case decode.JsonNodeTypeSlice:
			res[i] = marshalSlice(child)
		case decode.JsonNodeTypeObject:
			res[i] = marshalObject(child)
		}
	}
	return res
}

// Deprecated: 请使用 decode.UnMarshal
// Unmarshal 将使用 json 编码的数据反序列化为 JsonNode 对象。
func Unmarshal(input []byte) (*decode.JsonNode, error) {
	return Parse(input)
}

// Deprecated: 请使用 decode.Marshal
func Marshal(root *decode.JsonNode) ([]byte, error) {
	if root == nil {
		return nil, errors.New("can not marshal nil")
	}
	var dict interface{}
	switch root.Type {
	case decode.JsonNodeTypeObject:
		dict = marshalObject(root)
	case decode.JsonNodeTypeSlice:
		dict = marshalSlice(root)
	case decode.JsonNodeTypeValue:
		dict = marshalValue(root)
	}
	res, err := json.Marshal(dict)
	if err != nil {
		return nil, errors.Wrap(err, "fail to marshal")
	}
	return res, nil
}
