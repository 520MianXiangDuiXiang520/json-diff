package json_diff

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type JsonNodeType uint8

const (
	JsonNodeTypeValue = iota + 1
	JsonNodeTypeSlice
	JsonNodeTypeObject
)

type JsonNode struct {
	Type        JsonNodeType         `json:"type"`
	Hash        string               `json:"hash"`
	Key         string               `json:"key"`
	Value       interface{}          `json:"value"`        // 保存 JsonNodeTypeValue 类型对象的值
	Children    []*JsonNode          `json:"children"`     // 保存 JsonNodeTypeSlice 类型对象的值
	ChildrenMap map[string]*JsonNode `json:"children_map"` // 保存 JsonNodeTypeObject 类型对象的值
	Level       int64                `json:"level"`        // 该 node 所处的层级
}

func NewObjectNode(key string, childrenMap map[string]*JsonNode, level int) *JsonNode {
	return &JsonNode{
		Type:        JsonNodeTypeObject,
		Key:         key,
		ChildrenMap: childrenMap,
		Level:       int64(level),
	}
}

func NewSliceNode(children []*JsonNode, level int) *JsonNode {
	return &JsonNode{
		Type:     JsonNodeTypeSlice,
		Children: children,
		Level:    int64(level),
	}
}

func NewValueNode(value interface{}, level int) *JsonNode {
	return &JsonNode{
		Type:  JsonNodeTypeValue,
		Value: value,
		Level: int64(level),
	}
}

func (jn *JsonNode) ADD(key interface{}, value *JsonNode) error {
	switch jn.Type {
	case JsonNodeTypeObject:
		k, ok := key.(string)
		if !ok {
			return fmt.Errorf("you are trying to insert an object into a node of type JsonNodeTypeObject," +
				" please make sure that the type of key is string")
		}
		jn.ChildrenMap[k] = value
	case JsonNodeTypeSlice:
		ks, ok := key.(string)
		k, err := strconv.Atoi(ks)
		if !ok || err != nil {
			return fmt.Errorf("you are trying to insert an object into a node of type JsonNodeTypeSlice," +
				" please make sure that the type of key is int")
		}
		size := len(jn.Children)
		if k > size || k < 0 {
			// TODO: 当传入的 index 大于当前 Children 长度时，可以配置处理方式
			jn.Children = append(jn.Children, value)
		} else {
			n := make([]*JsonNode, size+1)
			for i := 0; i < k; i++ {
				n[i] = jn.Children[i]
			}
			n[k] = value
			for i := k + 1; i < size+1; i++ {
				n[i] = jn.Children[i-1]
			}
			jn.Children = n
		}
	default:
		return fmt.Errorf("cannot add an object to a node of type JsonNodeTypeValue")
	}
	return nil
}

// AddPath 为 node 的 path 路径处的对象添加一个子节点
// path 路径表示的是子节点加入后的路径, 以 "/" 开头
func AddPath(node *JsonNode, path string, value *JsonNode) error {
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return err
	}
	return f.ADD(childKey, value)
}

func (jn *JsonNode) String() string {
	return jn.Key
}

func (jn *JsonNode) Equal(patch *JsonNode) bool {
	if jn == nil && patch == nil {
		return true
	}
	if patch == nil || jn == nil {
		return false
	}
	if jn.Type != patch.Type {
		return false
	}

	switch jn.Type {
	case JsonNodeTypeSlice:
		if len(jn.Children) != len(patch.Children) {
			return false
		}
		for i, child := range jn.Children {
			if !child.Equal(patch.Children[i]) {
				return false
			}
		}
	case JsonNodeTypeObject:
		if len(jn.ChildrenMap) != len(patch.ChildrenMap) {
			return false
		}
		for key, value := range jn.ChildrenMap {
			v2, ok := patch.ChildrenMap[key]
			if !ok {
				return false
			}
			if !value.Equal(v2) {
				return false
			}
		}
	case JsonNodeTypeValue:
		return jn.Value == patch.Value
	}
	return true
}

func (jn *JsonNode) find(paths []string) (*JsonNode, bool) {
	root := jn
	for _, key := range paths {
		key = keyRestore(key)
		switch root.Type {
		case JsonNodeTypeObject:
			r, ok := root.ChildrenMap[key]
			if !ok {
				return nil, false
			}
			root = r
		case JsonNodeTypeSlice:
			n, err := strconv.Atoi(key)
			if err != nil {
				return nil, false
			}
			if n > len(root.Children)-1 {
				return nil, false
			}
			root = root.Children[n]
		case JsonNodeTypeValue:
			return root, true
		}
	}
	return root, true
}

// Find 根据路径返回对应的 node 节点
// 如完整的 json 文件为：
//
//   {
//    "article_list": [
//      {
//        "id": 1,
//        "article_info": {
//          "name": "瓦尔登湖",
//          "type": "文学"
//        },
//        "author_info": {
//          "name": "梭罗",
//          "country": "US"
//        }
//      },
//    ]
//   }
// 使用 `/article_list/0/author_info` 可以得到
//   {
//      "name": "梭罗",
//      "country": "US",
//   }
func (jn *JsonNode) Find(path string) (*JsonNode, bool) {
	link := strings.Split(path, "/")
	if len(link) <= 1 {
		return jn, true
	}
	return jn.find(link[1:])
}

var pathNotFind = errors.New("pathNotFind")

func (jn *JsonNode) Replace(key interface{}, value *JsonNode) (*JsonNode, error) {
	var old *JsonNode
	switch jn.Type {
	case JsonNodeTypeSlice:
		is, ok := key.(string)
		index, err := strconv.Atoi(is)
		if !ok || err != nil {
			return nil, fmt.Errorf("you are trying to replace a value from a node of type JsonNodeTypeSlice," +
				" please make sure that the key is of type int")
		}
		size := len(jn.Children)
		if index > size-1 || index < 0 {
			return nil, fmt.Errorf("index(%d) out of range (%d)", index, size)
		}
		old = jn.Children[index]
		jn.Children[index] = value
	case JsonNodeTypeObject:
		key, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("you are trying to replace a value from a node of type JsonNodeTypeObject," +
				" please make sure that the key is of type string")
		}
		old = jn.ChildrenMap[key]
		jn.ChildrenMap[key] = value
	case JsonNodeTypeValue:
		old = jn
		jn.Value = value.Value
	}
	return old, nil
}

// ReplacePath 替换 node 中 path 处的对象为 value, 并返回旧值
func ReplacePath(node *JsonNode, path string, value *JsonNode) (*JsonNode, error) {
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return nil, err
	}
	return f.Replace(childKey, value)
}

func (jn *JsonNode) Remove(key interface{}) (*JsonNode, error) {
	var old *JsonNode
	switch jn.Type {
	case JsonNodeTypeValue:
		return nil, fmt.Errorf("it is not allowed to execute Remove() on a node of type JsonNodeTypeValue." +
			" To delete a node of this type, please operate its parent node")
	case JsonNodeTypeObject:
		key, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("you are trying to remove a value from a node of type JsonNodeTypeObject," +
				" please make sure that the key is of type string")
		}
		if _, ok := jn.ChildrenMap[key]; !ok {
			return nil, fmt.Errorf("key(%s) does not exist", key)
		}
		old = jn.ChildrenMap[key]
		delete(jn.ChildrenMap, key)
	case JsonNodeTypeSlice:
		is, ok := key.(string)
		index, err := strconv.Atoi(is)
		if !ok || err != nil {
			return nil, fmt.Errorf("you are trying to remove a value from a node of type JsonNodeTypeSlice," +
				" please make sure that the key is of type int")
		}
		size := len(jn.Children)
		if index > size-1 || index < 0 {
			return nil, fmt.Errorf("index(%d) out of range (%d)", index, size)
		}
		old = jn.Children[index]
		n := make([]*JsonNode, size-1)
		for i := 0; i < index; i++ {
			n[i] = jn.Children[i]
		}
		for i := index + 1; i < size; i++ {
			n[i-1] = jn.Children[i]
		}
		jn.Children = n
	}
	return old, nil
}

func RemovePath(node *JsonNode, path string) (*JsonNode, error) {
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return nil, err
	}
	return f.Remove(childKey)
}

func MovePath(node *JsonNode, from, path string) (*JsonNode, error) {
	fromNode, ok := node.Find(from)
	if !ok {
		return nil, fmt.Errorf("from path(%s) not find", from)
	}
	_, ok = node.Find(path)
	if ok {
		_, err := ReplacePath(node, path, fromNode)
		if err != nil {
			return nil, err
		}
	} else {
		err := AddPath(node, path, fromNode)
		if err != nil {
			return nil, err
		}
	}
	old, err := RemovePath(node, from)
	if err != nil {
		return nil, err
	}
	return old, nil
}

func CopyPath(node *JsonNode, from, path string) error {
	fromNode, ok := node.Find(from)
	if !ok {
		return fmt.Errorf("from path(%s) not find", from)
	}
	return AddPath(node, path, fromNode)
}

func splitKey(node *JsonNode, path string) (string, *JsonNode, error) {
	paths := strings.Split(path, "/")[1:]
	size := len(paths)
	childKey := paths[size-1]
	paths = paths[:size-1]
	f, ok := node.find(paths)
	if !ok {
		return "", nil, fmt.Errorf("%s path not find", path)
	}
	return childKey, f, nil
}
