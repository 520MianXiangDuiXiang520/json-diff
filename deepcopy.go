package json_diff

import (
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
	"github.com/pkg/errors"
)

func copySlice(src *decode.JsonNode) (*decode.JsonNode, error) {
	size := len(src.Children)
	res := decode.NewSliceNode(make([]*decode.JsonNode, size), int(src.Level))
	for i, child := range src.Children {
		var newNode *decode.JsonNode
		var err error
		switch child.Type {
		case decode.JsonNodeTypeSlice:
			newNode, err = copySlice(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Slice type", i)
			}
		case decode.JsonNodeTypeObject:
			newNode, err = copyObject(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Object type", i)
			}
		case decode.JsonNodeTypeValue:
			newNode, err = copyValue(child)
			if err != nil {
				return nil, errors.Wrapf(err, "fail to copy %dst of Value type", i)
			}
		}
		res.Children[i] = newNode
	}
	return res, nil
}

func copyValue(src *decode.JsonNode) (*decode.JsonNode, error) {
	return decode.NewValueNode(src.Value, int(src.Level)), nil
}

func copyObject(src *decode.JsonNode) (*decode.JsonNode, error) {
	res := decode.NewObjectNode("", map[string]*decode.JsonNode{}, int(src.Level))
	for k, v := range src.ChildrenMap {
		var newNode *decode.JsonNode
		var err error
		switch v.Type {
		case decode.JsonNodeTypeObject:
			newNode, err = copyObject(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Object type", k)
			}
		case decode.JsonNodeTypeSlice:
			newNode, err = copySlice(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Slice type", k)
			}
		case decode.JsonNodeTypeValue:
			newNode, err = copyValue(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to copy %s of Value type", k)
			}
		}
		res.ChildrenMap[k] = newNode
	}
	return res, nil
}

func DeepCopy(src *decode.JsonNode) (*decode.JsonNode, error) {
	switch src.Type {
	case decode.JsonNodeTypeObject:
		return copyObject(src)
	case decode.JsonNodeTypeSlice:
		return copySlice(src)
	case decode.JsonNodeTypeValue:
		return copyValue(src)
	}
	return nil, errors.New("src has an unknown type")
}
