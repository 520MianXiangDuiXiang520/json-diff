package json_diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
)

func diffSlice(diffs *diffs, path string, source, patch *JsonNode, option JsonDiffOption) {
	lcsList := longestCommonSubsequence(source.Children, patch.Children)
	lcsIdx := 0
	srcIdx := 0
	tarIdx := 0
	pos := 0
	for lcsIdx < len(lcsList) {
		pathBuffer := bytes.NewBufferString(path)
		srcNode := source.Children[srcIdx]
		lcsNode := lcsList[lcsIdx]
		tarNode := patch.Children[tarIdx]
		if lcsNode.Equal(srcNode) && lcsNode.Equal(tarNode) {
			lcsIdx++
			srcIdx++
			tarIdx++
			pos++
		} else {
			if lcsNode.Equal(srcNode) {
				pathBuffer.WriteString("/")
				pathBuffer.WriteString(strconv.Itoa(pos))
				diffs.add(newDiffNode(DiffTypeAdd, pathBuffer.String(), tarNode, "", option))
				tarIdx++
				pos++
			} else if lcsNode.Equal(tarNode) {
				pathBuffer.WriteString("/")
				pathBuffer.WriteString(strconv.Itoa(pos))
				diffs.add(newDiffNode(DiffTypeRemove, pathBuffer.String(), srcNode, "", option))
				srcIdx++
			} else {
				pathBuffer.WriteString("/")
				pathBuffer.WriteString(strconv.Itoa(pos))
				diff(diffs, pathBuffer.String(), srcNode, tarNode, option)
				srcIdx++
				tarIdx++
				pos++
			}
		}
	}

	for srcIdx < len(source.Children) && tarIdx < len(patch.Children) {
		pathBuffer := bytes.NewBufferString(path)
		srcNode := source.Children[srcIdx]
		tarNode := patch.Children[tarIdx]
		pathBuffer.WriteString("/")
		pathBuffer.WriteString(strconv.Itoa(pos))
		diff(diffs, pathBuffer.String(), srcNode, tarNode, option)
		srcIdx++
		tarIdx++
		pos++
	}

	// 如果 source 或 patch 后面还有，属于 add 或 remove
	for ; srcIdx < len(source.Children); srcIdx++ {
		pathBuffer := bytes.NewBufferString(path)
		pathBuffer.WriteString("/")
		pathBuffer.WriteString(strconv.Itoa(pos))
		diffs.add(newDiffNode(DiffTypeRemove, pathBuffer.String(), source.Children[srcIdx], "", option))
		pos++
	}

	for ; tarIdx < len(patch.Children); tarIdx++ {
		pathBuffer := bytes.NewBufferString(path)
		pathBuffer.WriteString("/")
		pathBuffer.WriteString(strconv.Itoa(pos))
		diffs.add(newDiffNode(DiffTypeAdd, pathBuffer.String(), patch.Children[tarIdx], "", option))
		pos++
	}
}

func diffObject(diffs *diffs, path string, source, patch *JsonNode, option JsonDiffOption) {
	for srcKey, srcValue := range source.ChildrenMap {
		tarVal, tarOk := patch.ChildrenMap[srcKey]
		currPath := fmt.Sprintf("%s/%s", path, srcKey)
		if !tarOk {
			diffs.add(newDiffNode(DiffTypeRemove, currPath, srcValue, "", option))
			continue
		}
		diff(diffs, currPath, srcValue, tarVal, option)
	}

	for tarKey, tarVal := range patch.ChildrenMap {
		_, srcOk := source.ChildrenMap[tarKey]
		if !srcOk {
			currPath := fmt.Sprintf("%s/%s", path, tarKey)
			diffs.add(newDiffNode(DiffTypeAdd, currPath, tarVal, "", option))
		}
	}
}

func diff(diffs *diffs, path string, source, patch *JsonNode, option JsonDiffOption) {
	if source == nil && patch != nil {
		diffs.add(newDiffNode(DiffTypeAdd, path, patch, "", option))
	}
	if source != nil && patch == nil {
		diffs.add(newDiffNode(DiffTypeRemove, path, nil, "", option))
	}
	if source != nil && patch != nil {
		if source.Type == JsonNodeTypeObject && patch.Type == JsonNodeTypeObject {
			diffObject(diffs, path, source, patch, option)
		} else if source.Type == JsonNodeTypeSlice && patch.Type == JsonNodeTypeSlice {
			diffSlice(diffs, path, source, patch, option)
		} else {
			// 两个都是 JsonNodeTypeSlice
			if !source.Equal(patch) {
				diffs.add(newDiffNode(DiffTypeReplace, path, patch, "", option))
			}
		}
	}
}

// GetDiffNode 比较两个 JsonNode 之间的差异，并返回 JsonNode 格式的差异结果
func GetDiffNode(sourceJsonNode, patchJsonNode *JsonNode, options ...JsonDiffOption) *JsonNode {
	option := JsonDiffOption(0)
	for _, o := range options {
		option |= o
	}
	diffs := newDiffs()
	diff(diffs, "", sourceJsonNode, patchJsonNode, option)
	doOption(diffs, option, sourceJsonNode, patchJsonNode)
	return diffs.d
}

// AsDiffs 比较 patch 相比于 source 的差别，返回 json 格式的差异文档。
func AsDiffs(source, patch []byte, options ...JsonDiffOption) ([]byte, error) {
	sourceJsonNode, _ := Unmarshal(source)
	patchJsonNode, _ := Unmarshal(patch)
	dict := marshalSlice(GetDiffNode(sourceJsonNode, patchJsonNode, options...))
	return json.Marshal(dict)
}

func merge(srcNode, diffNode *JsonNode) error {
	for _, diff := range diffNode.Children {
		if diff.Type != JsonNodeTypeObject {
			return errors.WithStack(BadDiffsError)
		}
		op := diff.ChildrenMap["op"].Value
		path := diff.ChildrenMap["path"].Value.(string)
		switch op {
		case "add":
			err := AddPath(srcNode, path, diff.ChildrenMap["value"])
			if err != nil {
				return err
			}
		case "remove":
			_, err := RemovePath(srcNode, path)
			if err != nil {
				return err
			}
		case "replace":
			val := diff.ChildrenMap["value"]
			_, err := ReplacePath(srcNode, path, val)
			if err != nil {
				return err
			}
		case "move":
			from := diff.ChildrenMap["from"].Value.(string)
			_, err := MovePath(srcNode, from, path)
			if err != nil {
				return err
			}
		case "copy":
			from := diff.ChildrenMap["from"].Value.(string)
			err := CopyPath(srcNode, from, path)
			if err != nil {
				return err
			}
		case "test":
			err := ATestPath(srcNode, path, diff.ChildrenMap["value"])
			if err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("bad diffs: %v", diff))
		}
	}
	return nil
}

// MergeDiff 根据差异文档 diff 还原 source 的差异
func MergeDiff(source, diff []byte) ([]byte, error) {
	diffNode, err := Unmarshal(diff)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal diff data")
	}
	srcNode, err := Unmarshal(source)
	if err != nil {
		return nil, errors.Wrap(err, "fail to unmarshal source data")
	}
	result, err := MergeDiffNode(srcNode, diffNode)
	if err != nil {
		return nil, errors.Wrap(err, "fail to merge diff")
	}
	return Marshal(result)
}

// MergeDiffNode 将 JsonNode 类型的 diffs 应用于源 source 上，并返回合并后的新 jsonNode 对象
// 如果 diffs 不合法，第二个参数将会返回 BadDiffsError
func MergeDiffNode(source, diffs *JsonNode) (*JsonNode, error) {
	if diffs == nil {
		return source, nil
	}
	if diffs.Type != JsonNodeTypeSlice {
		return nil, errors.New("bad diffs")
	}
	copyNode, err := DeepCopy(source)
	if err != nil {
		return nil, errors.Wrap(err, "fail to deep copy source")
	}
	err = merge(copyNode, diffs)
	if err != nil {
		return nil, errors.Wrap(err, "fail to merge")
	}
	return copyNode, nil
}
