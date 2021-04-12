package json_diff

type DiffType int

// RFC6902 https://tools.ietf.org/html/rfc6902
const (
	DiffTypeAdd = iota + 1
	DiffTypeRemove
	DiffTypeReplace
	DiffTypeMove
	DiffTypeCopy
	DiffTypeTest
)

func translateDiffType(t DiffType) string {
	switch t {
	case DiffTypeAdd:
		return "add"
	case DiffTypeCopy:
		return "copy"
	case DiffTypeMove:
		return "move"
	case DiffTypeRemove:
		return "remove"
	case DiffTypeReplace:
		return "replace"
	case DiffTypeTest:
		return "test"
	}
	return ""
}

var diffTypeTable = map[string]DiffType{
	"add":     DiffTypeAdd,
	"remove":  DiffTypeRemove,
	"move":    DiffTypeMove,
	"replace": DiffTypeReplace,
	"copy":    DiffTypeCopy,
	"test":    DiffTypeTest,
}

func stringToDiffType(s string) (DiffType, bool) {
	v, ok := diffTypeTable[s]
	return v, ok
}

func newDiffNode(diffType DiffType, path string, value *JsonNode, from string, opt JsonDiffOption) *JsonNode {
	n := &JsonNode{
		Type:        JsonNodeTypeObject,
		ChildrenMap: make(map[string]*JsonNode),
	}
	_ = n.ADD("op", NewValueNode(translateDiffType(diffType), 1))
	_ = n.ADD("path", NewValueNode(path, 1))
	switch diffType {
	case DiffTypeAdd, DiffTypeTest, DiffTypeReplace:
		_ = n.ADD("value", value)
	case DiffTypeMove, DiffTypeCopy:
		_ = n.ADD("from", NewValueNode(from, 1))
	case DiffTypeRemove:
		if opt&UseFullRemoveOption == UseFullRemoveOption {
			_ = n.ADD("value", value)
		}
	}
	return n
}
