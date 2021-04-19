package json_diff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type diffs struct {
	d *JsonNode
}

func (d *diffs) remove(idx int) {
	if idx >= d.size() {
		return
	}
	n := make([]*JsonNode, d.size()-1)
	for i := 0; i < idx; i++ {
		n[i] = d.d.Children[i]
	}
	for i := idx + 1; i < d.size(); i++ {
		n[i-1] = d.d.Children[i]
	}
	d.d.Children = n
}

func (d *diffs) get(i int) *JsonNode {
	if i >= d.size() {
		return nil
	}
	return d.d.Children[i]
}

func (d *diffs) add(node *JsonNode) {
	d.d.Children = append(d.d.Children, node)
}

func (d *diffs) insert(i int, node *JsonNode) {
	if i > d.size() {
		d.add(node)
		return
	}
	n := make([]*JsonNode, d.size()+1)
	for idx := 0; idx < i; idx++ {
		n[idx] = d.d.Children[idx]
	}
	n[i] = node
	for idx := i; idx < d.size(); idx++ {
		n[idx+1] = d.d.Children[idx]
	}
	d.d.Children = n
}

func (d *diffs) set(i int, node *JsonNode) {
	if i < len(d.d.Children) {
		d.d.Children[i] = node
	}
}

func (d *diffs) size() int {
	return len(d.d.Children)
}

func (d *diffs) rangeType(f func(i int, v *JsonNode, t DiffType) bool) {
	for i, child := range d.d.Children {
		ty := child.ChildrenMap["op"].Value.(string)
		t, ok := stringToDiffType(ty)
		if !ok {
			continue
		}
		if f(i, child, t) {
			break
		}
	}
}

func (d *diffs) ranger(f func(i int, v *JsonNode) bool) {
	for i, child := range d.d.Children {
		if f(i, child) {
			break
		}
	}
}

func newDiffs() *diffs {
	return &diffs{d: &JsonNode{
		Type: JsonNodeTypeSlice,
	}}
}

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
		} else if source.Type == JsonNodeTypeSlice || patch.Type == JsonNodeTypeSlice {
			diffs.add(newDiffNode(DiffTypeReplace, path, patch, "", option))
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
    sourceJsonNode := Unmarshal(source)
    patchJsonNode := Unmarshal(patch)
	dict := marshalSlice(GetDiffNode(sourceJsonNode, patchJsonNode, options...))
	return json.Marshal(dict)
}

var badDiff = errors.New("DabDiff")

func merge(srcNode, diffNode *JsonNode) error {
	for _, diff := range diffNode.Children {
		if diff.Type != JsonNodeTypeObject {
			return badDiff
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
			err := mergeTest(srcNode, path, diff.ChildrenMap["value"])
			if err != nil {
				return err
			}
		default:
			return badDiff
		}
	}
	return nil
}

func testFail(info string) error {
	return fmt.Errorf("TestFail: %s", info)
}

func mergeTest(srcNode *JsonNode, path string, value *JsonNode) error {
	f, ok := srcNode.Find(path)
	if !ok {
		return testFail("path not find")
	}
	if f.Type != f.Type {
		return testFail("different types")
	}
	switch value.Type {
	case JsonNodeTypeValue:
		// [{"op": "test", "path": "a/b/c", "value":"123"}]
		if f.Value != value.Value {
			return testFail("different value")
		}
	case JsonNodeTypeSlice:
		// [{"op": "test", "path": "a/b/c", "value":[123, 456]}]
		if len(f.Children) != len(value.Children) {
			return testFail("different value")
		}
		for i, v := range value.Children {
			if !v.Equal(f.Children[i]) {
				return testFail("different value")
			}
		}
	case JsonNodeTypeObject:
		if len(f.ChildrenMap) != len(value.ChildrenMap) {
			return testFail("different value")
		}
		for k, v := range value.ChildrenMap {
			if !v.Equal(f.ChildrenMap[k]) {
				return testFail("different value")
			}
		}
	}
	return nil
}

// 根据差异文档 diff 还原 source 的差异
func MergeDiff(source, diff []byte) ([]byte, error) {
	diffNode := Unmarshal(diff)
	srcNode := Unmarshal(source)
	result, err := MergeDiffNode(srcNode, diffNode)
	if err != nil {
	    return nil, err
    }
	return Marshal(result)
}

func MergeDiffNode(source, diffs *JsonNode) (*JsonNode, error) {
    if diffs == nil {
        return source, nil
    }
    if diffs.Type != JsonNodeTypeSlice {
        return nil, errors.New("DabDiff")
    }
    copyNode := new(JsonNode)
    err := DeepCopy(copyNode, source)
    if err != nil {
        return nil, err
    }
    err = merge(copyNode, diffs)
    if err != nil {
        return nil, err
    }
    source = copyNode
    return source, nil
}

func beautifyJsonString(data []byte) {
	var str bytes.Buffer
	_ = json.Indent(&str, data, "", "    ")
	fmt.Println(str.String())
}
