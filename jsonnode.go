package json_diff

import (
	"fmt"
	"strconv"
	"strings"
)

type JsonNodeType uint8

const (
	// JsonNodeTypeValue :  普通值类型，对应 int, float, string, bool 等，
	// 该类型不具有子类型，存储在 Value 字段；
	JsonNodeTypeValue = iota + 1

	// JsonNodeTypeSlice :  切片类型，对应 []，该类型是有序的，
	// 存储在 Children 字段，使用下标唯一表示；
	JsonNodeTypeSlice

	// JsonNodeTypeObject : 对象类型，对应 {}，该类型是无序的，
	// 存储在 ChildrenMap， 使用 key 唯一表示。
	JsonNodeTypeObject
)

func (jt JsonNodeType) String() string {
    switch jt {
    case JsonNodeTypeValue:
        return "value"
    case JsonNodeTypeSlice:
        return "slice"
    case JsonNodeTypeObject:
        return "object"
    }
    return ""
}

// JsonNode 以树的形式组织 Json 中的每一项数据。
// 根据 Json 的特点，可以将 Json 存储的数据分为三种不同类型:
// JsonNodeTypeValue，JsonNodeTypeSlice，JsonNodeTypeObject
// 如：
//    {
//      "a": 1,
//      "b": [1],
//    }
// 就可以看作两个 JsonNodeTypeObject 类型节点，key 分别是 a 和 b,
// 其中 a 的值是一个值为 1 的 JsonNodeTypeValue，
// b 的值是一个长度为 1 的 JsonNodeTypeSlice 类型节点，
// 而他第 0 个元素也是一个值为 1 的 JsonNodeTypeValue 节点。
// 最外层的大括号是一个 JsonNodeTypeObject 节点，他作为根节点，Key 为空。
//
// 一个 Json 字节数组可以使用 Unmarshal 反序列化为 JsonNode 对象，
// JsonNode 对象也可以使用 Marshal 序列化为 Json 字节数组
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

// ADD 为当前的 JsonNode 节点添加子对象。
// 当当前节点为 JsonNodeTypeObject 类型时，key 必须是 string 类型；
// 当当前节点为 JsonNodeTypeSlice 类型时，key 表示新加入节点的位置下标，必须能转换为 int 类型；
// 不能对 JsonNodeTypeValue 类型的节点执行 ADD 操作；
// 不符合上述要求该方法返回一个由 BadDiffsError 装饰的 error
func (jn *JsonNode) ADD(key interface{}, value *JsonNode) error {
	switch jn.Type {
	case JsonNodeTypeObject:
		k, ok := key.(string)
		if !ok {
			return GetJsonNodeError("add", keyMastString(key))
		}
		jn.ChildrenMap[k] = value
	case JsonNodeTypeSlice:
		k := 0
		switch key.(type) {
		case string:
			var err error
			k, err = strconv.Atoi(key.(string))
			if err != nil {
				return GetJsonNodeError("add", keyMustCanBeConvertibleToInt(key))
			}
		case int:
			k = key.(int)
		default:
			return GetJsonNodeError("add", keyMustCanBeConvertibleToInt(key))
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
		return GetJsonNodeError("add",
			"cannot add an object to a node of type JsonNodeTypeValue")
	}
	return nil
}

// AddPath 为 node 的 path 路径处的对象添加一个子节点
// path 路径表示的是子节点加入后的路径, 以 "/" 开头
func AddPath(node *JsonNode, path string, value *JsonNode) error {
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return GetJsonNodeError("add",
			fmt.Sprintf("path (%s) is not compliant", path))
	}
	return f.ADD(childKey, value)
}

func (jn *JsonNode) String() string {
	return jn.Key
}

// Equal 比较当前节点和 patch 是否相等
// 对于两个 JsonNodeTypeObject 类型，不关心顺序，每个 key 对应的 value 都相等才认为相等；
// 对于两个 JsonNodeTypeSlice 类型，Children 中每个位置对应的元素都相等才认为相等；
// 对于两个 JsonNodeTypeValue 类型，Value 相等即认为相等。
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

// Replace 使用 value 替换当前节点的 key 的值, 并返回旧值。
// 当当前节点为 JsonNodeTypeObject 类型时，key 必须是 string 类型；
// 当当前节点为 JsonNodeTypeSlice 类型时，key 表示新加入节点的位置下标，必须能转换为 int 类型；
// 不符合上述要求该方法返回一个由 BadDiffsError 装饰的 error
func (jn *JsonNode) Replace(key interface{}, value *JsonNode) (*JsonNode, error) {
	var old *JsonNode
	switch jn.Type {
	case JsonNodeTypeSlice:
		index := 0
		var err error
		switch key.(type) {
		case string:
			index, err = strconv.Atoi(key.(string))
			if err != nil {
				return nil, GetJsonNodeError("replace", keyMustCanBeConvertibleToInt(key))
			}
		case int:
			index = key.(int)
		default:
			return nil, GetJsonNodeError("replace", keyMustCanBeConvertibleToInt(key))
		}
		size := len(jn.Children)
		if index > size-1 || index < 0 {
			return nil, GetJsonNodeError("replace",
				fmt.Sprintf("index(%d) out of range (%d)", index, size))
		}
		old = jn.Children[index]
		jn.Children[index] = value
	case JsonNodeTypeObject:
		key, ok := key.(string)
		if !ok {
			return nil, GetJsonNodeError("replace", keyMastString(key))
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
	// 游标移动到 path 对应的位置
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return nil, GetJsonNodeError("replace", fmt.Sprintf("path (%s) is not compliant", path))
	}
	return f.Replace(childKey, value)
}

// Remove 删除当前 JsonNode 中 key 对应的节点并返回被删除的值。
// Remove 只能删除父节点的某个子节点，节点不能删除它自己，因此，
// JsonNodeTypeValue 类型的节点不能使用 Remove 方法。
func (jn *JsonNode) Remove(key interface{}) (*JsonNode, error) {
	var old *JsonNode
	switch jn.Type {
	case JsonNodeTypeValue:
		return nil, GetJsonNodeError("remove", "unable to execute remove on JsonNodeTypeValue")
	case JsonNodeTypeObject:
		key, ok := key.(string)
		if !ok {
			return nil, GetJsonNodeError("remove", keyMastString(key))
		}
		if _, ok := jn.ChildrenMap[key]; !ok {
			return nil, GetJsonNodeError("remove", fmt.Sprintf("key(%s) does not exist", key))
		}
		old = jn.ChildrenMap[key]
		delete(jn.ChildrenMap, key)
	case JsonNodeTypeSlice:
		index := 0
		var err error
		switch key.(type) {
		case string:
			index, err = strconv.Atoi(key.(string))
			if err != nil {
				return nil, GetJsonNodeError("remove", keyMustCanBeConvertibleToInt(key))
			}
		case int:
			index = key.(int)
		default:
			return nil, GetJsonNodeError("remove", keyMustCanBeConvertibleToInt(key))
		}
		size := len(jn.Children)
		if index > size-1 || index < 0 {
			return nil, GetJsonNodeError("remove", fmt.Sprintf("index(%d) out of range (%d)", index, size))
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

// RemovePath 删除并返回 node 中根据 path 找到的节点。
func RemovePath(node *JsonNode, path string) (*JsonNode, error) {
	childKey, f, err := splitKey(node, path)
	if err != nil {
		return nil, GetJsonNodeError("remove", fmt.Sprintf("path (%s) is not compliant", path))
	}
	return f.Remove(childKey)
}

// MovePath 将 node 中 from 处的节点移动到 path 处
func MovePath(node *JsonNode, from, path string) (*JsonNode, error) {
	fromNode, ok := node.Find(from)
	if !ok {
		return nil, GetJsonNodeError("move", fmt.Sprintf("from path(%s) not find", from))
	}
	_, ok = node.Find(path)
	if ok {
		_, err := ReplacePath(node, path, fromNode)
		if err != nil {
			return nil, WrapJsonNodeError("move", err)
		}
	} else {
		err := AddPath(node, path, fromNode)
		if err != nil {
			return nil, WrapJsonNodeError("move", err)
		}
	}
	old, err := RemovePath(node, from)
	if err != nil {
		return nil, WrapJsonNodeError("move", err)
	}
	return old, nil
}

// CopyPath 将 node from 处的节点复制到 path 处
func CopyPath(node *JsonNode, from, path string) error {
	fromNode, ok := node.Find(from)
	if !ok {
		return GetJsonNodeError("copy", fmt.Sprintf("from path(%s) not find", from))
	}
	err := AddPath(node, path, fromNode)
	if err != nil {
		return WrapJsonNodeError("copy", err)
	}
	return nil
}

func ATestPath(srcNode *JsonNode, path string, value *JsonNode) error {
    f, ok := srcNode.Find(path)
    if !ok {
        return GetJsonNodeError("test", fmt.Sprintf("%s not find", path))
    }
    if f.Type != value.Type {
        return GetJsonNodeError("test",
            fmt.Sprintf("types are not equal, one is %s, another is %s",
                f.Type.String(), value.Type.String()))
    }
    switch value.Type {
    case JsonNodeTypeValue:
        // [{"op": "test", "path": "a/b/c", "value":"123"}]
        if f.Value != value.Value {
            return GetJsonNodeError("test", valueAreNotEqual(f.Value, value.Value))
        }
    case JsonNodeTypeSlice:
        // [{"op": "test", "path": "a/b/c", "value":[123, 456]}]
        if len(f.Children) != len(value.Children) {
            return GetJsonNodeError("test", valueAreNotEqual(f.Children, value.Children))
        }
        for i, v := range value.Children {
            if !v.Equal(f.Children[i]) {
                return GetJsonNodeError("test", valueAreNotEqual(v, f.Children[i]))
            }
        }
    case JsonNodeTypeObject:
        if len(f.ChildrenMap) != len(value.ChildrenMap) {
            return GetJsonNodeError("test", valueAreNotEqual(f.ChildrenMap, value.ChildrenMap))
        }
        for k, v := range value.ChildrenMap {
            if !v.Equal(f.ChildrenMap[k]) {
                return GetJsonNodeError("test", valueAreNotEqual(v, f.ChildrenMap[k]))
            }
        }
    }
    return nil
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

func keyMastString(got interface{}) string {
	return fmt.Sprintf("the key of JsonNodeTypeObject must be a string, got %v", got)
}

func keyMustCanBeConvertibleToInt(got interface{}) string {
	return fmt.Sprintf("the key of JsonNodeTypeSlice must be convertible to int, got %v", got)
}

func valueAreNotEqual(one, another interface{}) string {
    return fmt.Sprintf("value are not equal, one is %v, another is %v", one, another)
}
