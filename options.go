package json_diff

import (
	"fmt"
)

type JsonDiffOption uint

const (
	// 返回差异时使用 Copy, 当发现新增的子串出现在原串中时，使用该选项可以将 Add 行为替换为 Copy 行为
	// 以减少差异串的大小，但这需要额外的计算，默认不开启
	UseCopyOption JsonDiffOption = 1 << iota

	// 仅在 UseCopyOption 选项开启时有效，替换前会添加 Test 行为，以确保 Copy 的路径存在
	UseCheckCopyOption

	// 返回差异时使用 Copy, 当发现差异串中两个 Add 和 Remove 的值相等时，会将他们合并为一个 Move 行为
	// 以此减少差异串的大小，默认不开启
	UseMoveOption

	// Remove 时除了返回 path, 还返回删除了的值，默认不开启
	UseFullRemoveOption
)

func doOption(diffs *diffs, opt JsonDiffOption, src, target *JsonNode) {
	if diffs.d.Type != JsonNodeTypeSlice {
		return
	}
	if opt&UseCopyOption == UseCopyOption || opt&UseMoveOption == UseMoveOption {
		_ = setHash(src)
		_ = setHash(target)
	}
	if opt&UseCopyOption == UseCopyOption {
		doCopyOption(diffs, opt, src, target)
	}
	if opt&UseMoveOption == UseMoveOption {
		doMoveOption(diffs, opt, src, target)
	}
}

func doCopyOption(diffs *diffs, opt JsonDiffOption, src, target *JsonNode) {
	unChanges := getUnChangeNodes(src, target)
	diffs.rangeType(func(i int, v *JsonNode, t DiffType) bool {
		if t == DiffTypeAdd {
			key := setHash(v.ChildrenMap["value"])
			if path, ok := unChanges.load(key, v.ChildrenMap["value"]); ok {
				if opt&UseCheckCopyOption == 1 {
					diffs.insert(i, newDiffNode(DiffTypeTest, path.path.(string), v, "", opt))
					i++
				}
				diffs.set(i, newDiffNode(DiffTypeCopy, v.ChildrenMap["path"].Value.(string), nil, path.path.(string), opt))
			}
		}
		return false
	})
}

// func doMoveOption(diffs *diffs, opt JsonDiffOption, src, target *JsonNode) {
//     for index, diff := range diffs.d.Children {
//         t := diff.ChildrenMap["op"].Value.(string)
//         ty, _ := stringToDiffType(t)
//         switch ty {
//         case DiffTypeRemove:
//             removerPath := diff.ChildrenMap["path"].Value.(string)
//             value, _ := src.Find(removerPath)
//             for i := index + 1; i < diffs.size(); i++ {
//                 diff2 := diffs.get(i)
//                 ty2, _ := stringToDiffType(diff2.ChildrenMap["op"].Value.(string))
//                 if ty2 == DiffTypeAdd {
//                     value2 := diff2.ChildrenMap["value"]
//                     if value.Equal(value2) {
//                         path := diff2.ChildrenMap["path"].Value.(string)
//                         diffs.remove(index)
//                         diffs.set(i, newDiffNode(DiffTypeMove, path, nil, removerPath, opt))
//                     }
//                 }
//             }
//         case DiffTypeAdd:
//             value := diff.ChildrenMap["value"]
//             path := diff.ChildrenMap["path"].Value.(string)
//             for i := index + 1; i < diffs.size(); i ++ {
//                 diff2 := diffs.get(i)
//                 ty2, _ := stringToDiffType(diff2.ChildrenMap["op"].Value.(string))
//                 if ty2 == DiffTypeRemove {
//                     removerPath := diff2.ChildrenMap["path"].Value.(string)
//                     value2, _ := src.Find(removerPath)
//                     if value2.Equal(value) {
//                         diffs.remove(i)
//                         diffs.set(index, newDiffNode(DiffTypeMove, path, nil, removerPath, opt))
//                     }
//                 }
//             }
//         }
//
//     }
// }

func getDiffType(diff *JsonNode) DiffType {
	t, _ := stringToDiffType(diff.ChildrenMap["op"].Value.(string))
	return t
}

func getDiffValue(src, diff *JsonNode) *JsonNode {
	switch getDiffType(diff) {
	case DiffTypeAdd:
		return diff.ChildrenMap["value"]
	case DiffTypeRemove:
		p := diff.ChildrenMap["path"].Value.(string)
		v, _ := src.Find(p)
		return v
	}
	return nil
}

func getDiffPath(diff *JsonNode) string {
	return diff.ChildrenMap["path"].Value.(string)
}

func doMoveOption(diffs *diffs, opt JsonDiffOption, src, target *JsonNode) {
	for i := 0; i < diffs.size(); i++ {
		diff1 := diffs.get(i)
		diff1Type := getDiffType(diff1)
		if !(diff1Type == DiffTypeRemove || diff1Type == DiffTypeAdd) {
			continue
		}
		for j := i + 1; j < diffs.size(); j++ {
			diff2 := diffs.get(j)
			if !getDiffValue(src, diff1).Equal(getDiffValue(src, diff2)) {
				continue
			}
			var moveDiff *JsonNode
			if getDiffType(diff1) == DiffTypeRemove && getDiffType(diff2) == DiffTypeAdd {
				moveDiff = newDiffNode(DiffTypeMove, getDiffPath(diff2), nil, getDiffPath(diff1), opt)
			} else if getDiffType(diff1) == DiffTypeAdd && getDiffType(diff2) == DiffTypeRemove {
				moveDiff = newDiffNode(DiffTypeMove, getDiffPath(diff1), nil, getDiffPath(diff2), opt)
			}
			if moveDiff != nil {
				diffs.remove(j)
				diffs.set(i, moveDiff)
				break
			}
		}
	}
}

type unChangeContainerValue struct {
	path interface{}
	node *JsonNode
}

type unChangeContainer struct {
	c map[string][]*unChangeContainerValue
}

func (u *unChangeContainer) store(key string, obj *JsonNode, path interface{}) {
	list, ok := u.c[key]
	if !ok {
		u.c[key] = []*unChangeContainerValue{
			{path: path, node: obj},
		}
		return
	}
	list = append(list, &unChangeContainerValue{node: obj, path: path})
	u.c[key] = list
}

func (u *unChangeContainer) load(key string, obj *JsonNode) (*unChangeContainerValue, bool) {
	list, ok := u.c[key]
	if !ok {
		return nil, false
	}
	for _, v := range list {
		if v.node.Equal(obj) {
			return v, true
		}
	}
	return nil, false
}

func (u *unChangeContainer) storeOrLoad(key, path string, obj *JsonNode) (interface{}, bool) {
	p, ok := u.load(key, obj)
	if ok {
		return p, false
	}
	u.store(key, obj, path)
	return path, true
}

func getUnChangeNodes(src *JsonNode, target *JsonNode) unChangeContainer {
	contains := unChangeContainer{
		c: make(map[string][]*unChangeContainerValue),
	}
	computeUnChangeNode(&contains, "", src, target)
	return contains
}

func computeUnChangeNode(container *unChangeContainer, path string, src, target *JsonNode) {
	if src.Equal(target) {
		_, _ = container.storeOrLoad(src.Hash, path, src)
		return
	}
	if src.Type == target.Type {
		switch src.Type {
		case JsonNodeTypeObject:
			computeObjectUnChange(container, path, src, target)
		case JsonNodeTypeSlice:
			computeSliceUnChange(container, path, src, target)
		}
	}
}

func computeSliceUnChange(contains *unChangeContainer, path string, src, target *JsonNode) {
	for i, v := range src.Children {
		computeUnChangeNode(contains, fmt.Sprintf("%s/%d", path, i), v, target.Children[i])
	}
}

func computeObjectUnChange(contains *unChangeContainer, path string, src, target *JsonNode) {
	for k, v := range src.ChildrenMap {
		if tarV, ok := target.ChildrenMap[k]; ok {
			computeUnChangeNode(contains, fmt.Sprintf("%s/%s", path, k), v, tarV)
		}
	}
}
