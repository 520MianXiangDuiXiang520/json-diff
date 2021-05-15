package json_diff

import (
	"github.com/pkg/errors"
)

func copySlice(src *JsonNode) (*JsonNode, error) {
	size := len(src.Children)
	res := NewSliceNode(make([]*JsonNode, size), int(src.Level))
	for i, child := range src.Children {
		var newNode *JsonNode
		var err error
		switch child.Type {
		case JsonNodeTypeSlice:
			newNode, err = copySlice(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Slice type", i)
			}
		case JsonNodeTypeObject:
			newNode, err = copyObject(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Object type", i)
			}
		case JsonNodeTypeValue:
			newNode, err = copyValue(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Value type", i)
			}
		}
		res.Children[i] = newNode
	}
	return res, nil
}

func copyValue(src *JsonNode) (*JsonNode, error) {
	return NewValueNode(src.Value, int(src.Level)), nil
}

func copyObject(src *JsonNode) (*JsonNode, error) {
	res := NewObjectNode("", map[string]*JsonNode{}, int(src.Level))
	for k, v := range src.ChildrenMap {
		var newNode *JsonNode
		var err error
		switch v.Type {
		case JsonNodeTypeObject:
			newNode, err = copyObject(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Object type", k)
			}
		case JsonNodeTypeSlice:
			newNode, err = copySlice(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Slice type", k)
			}
		case JsonNodeTypeValue:
			newNode, err = copyValue(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Value type", k)
			}
		}
		res.ChildrenMap[k] = newNode
	}
	return res, nil
}

func DeepCopy(src *JsonNode) (*JsonNode, error) {
	switch src.Type {
	case JsonNodeTypeObject:
		return copyObject(src)
	case JsonNodeTypeSlice:
		return copySlice(src)
	case JsonNodeTypeValue:
		return copyValue(src)
	}
	return nil, errors.New("src has an unknown type")
}
