package json_diff

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

var keyReplaceRegexp = regexp.MustCompile(`~0*1`)

// key 中的
// "/"    会被替换成 "~1"
// "~1"   会被替换成 "~01"
// "~01"  会被替换为 "~001"
// "~001" 会被替换为 "~0001"
// 依此类推
func keyReplace(key string) string {
	resList := keyReplaceRegexp.FindAllStringIndex(key, -1)
	buff := bytes.NewBufferString("")
	pre := 0
	for _, res := range resList {
		buff.WriteString(key[pre:res[0]])
		buff.WriteRune('~')
		for i := 1; i < res[1]-res[0]; i++ {
			buff.WriteRune('0')
		}
		buff.WriteRune('1')
		pre = res[1]
	}
	buff.WriteString(key[pre:])
	return strings.ReplaceAll(buff.String(), "/", "~1")
}

func keyRestore(key string) string {
	key = strings.ReplaceAll(key, "~1", "/")
	resList := keyReplaceRegexp.FindAllStringIndex(key, -1)
	buff := bytes.NewBufferString("")
	pre := 0
	for _, res := range resList {
		buff.WriteString(key[pre:res[0]])
		buff.WriteRune('~')
		for i := 3; i < res[1]-res[0]; i++ {
			buff.WriteRune('0')
		}
		buff.WriteRune('1')
		pre = res[1]
	}
	buff.WriteString(key[pre:])
	return buff.String()
}

func parse(v interface{}, level int64) *JsonNode {
	var root *JsonNode
	switch v.(type) {
	case map[string]interface{}:
		value := v.(map[string]interface{})
		root = &JsonNode{Type: JsonNodeTypeObject, Level: level, ChildrenMap: make(map[string]*JsonNode)}
		for key, va := range value {
			key = keyReplace(key)
			n := parse(va, level+1)
			n.Key = key
			// n.Father = root
			root.ChildrenMap[key] = n
			// root.Children = append(root.Children, n)
		}
	case []interface{}:
		root = &JsonNode{Type: JsonNodeTypeSlice, Level: level}
		value := v.([]interface{})
		for _, va := range value {
			n := parse(va, level+1)
			// n.Father = root
			root.Children = append(root.Children, n)
		}
	default:
		root = &JsonNode{Type: JsonNodeTypeValue, Level: level}
		root.Type = JsonNodeTypeValue
		root.Value = v
	}
	return root
}

func Parse(input []byte) *JsonNode {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		panic(err)
	}
	return parse(v, 0)
}

func marshalValue(root *JsonNode) interface{} {
	return root.Value
}

func marshalObject(root *JsonNode) map[string]interface{} {
	dict := make(map[string]interface{})
	for k, v := range root.ChildrenMap {
		switch v.Type {
		case JsonNodeTypeObject:
			dict[k] = marshalObject(v)
		case JsonNodeTypeSlice:
			dict[k] = marshalSlice(v)
		case JsonNodeTypeValue:
			dict[k] = marshalValue(v)
		}
	}
	return dict
}

func marshalSlice(root *JsonNode) []interface{} {
	res := make([]interface{}, len(root.Children))
	for i, child := range root.Children {
		switch child.Type {
		case JsonNodeTypeValue:
			res[i] = marshalValue(child)
		case JsonNodeTypeSlice:
			res[i] = marshalSlice(child)
		case JsonNodeTypeObject:
			res[i] = marshalObject(child)
		}
	}
	return res
}

func Unmarshal(input []byte) *JsonNode {
	return Parse(input)
}

func Marshal(root *JsonNode) ([]byte, error) {
	var dict interface{}
	switch root.Type {
	case JsonNodeTypeObject:
		dict = marshalObject(root)
	case JsonNodeTypeSlice:
		dict = marshalSlice(root)
	case JsonNodeTypeValue:
		dict = marshalValue(root)
	}
	return json.Marshal(dict)

}
