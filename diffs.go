/*
 * Copyright 2021 Junebao
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package json_diff

import (
	"github.com/520MianXiangDuiXiang520/json-diff/decode"
)

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

func (dt DiffType) String() string {
	switch dt {
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

func newDiffNode(diffType DiffType, path string, value *decode.JsonNode, from string, opt JsonDiffOption) *decode.JsonNode {
	n := &decode.JsonNode{
		Type:        decode.JsonNodeTypeObject,
		ChildrenMap: make(map[string]*decode.JsonNode),
	}
	_ = n.ADD("op", decode.NewValueNode(diffType.String(), 1))
	_ = n.ADD("path", decode.NewValueNode(path, 1))
	switch diffType {
	case DiffTypeAdd, DiffTypeTest, DiffTypeReplace:
		_ = n.ADD("value", value)
	case DiffTypeMove, DiffTypeCopy:
		_ = n.ADD("from", decode.NewValueNode(from, 1))
	case DiffTypeRemove:
		if opt&UseFullRemoveOption == UseFullRemoveOption {
			_ = n.ADD("value", value)
		}
	}
	return n
}

// diffs 用来表示两个 JsonNode 之间的差异，其本身也是 JsonNode 的 Slice 类型
// 每一条差异保存在 diffs.d.Children 中
type diffs struct {
	d *decode.JsonNode
}

// remove 从 diffs.d.Children 中删除下标为 idx 的差异记录
func (d *diffs) remove(idx int) {
	if idx >= d.size() {
		return
	}
	n := make([]*decode.JsonNode, d.size()-1)
	for i := 0; i < idx; i++ {
		n[i] = d.d.Children[i]
	}
	for i := idx + 1; i < d.size(); i++ {
		n[i-1] = d.d.Children[i]
	}
	d.d.Children = n
}

// get 从 diffs.d.Children 中返回下表为 i 的差异记录
func (d *diffs) get(i int) *decode.JsonNode {
	if i >= d.size() {
		return nil
	}
	return d.d.Children[i]
}

// add 往 diffs.d.Children 尾部插入一条新的差异记录
func (d *diffs) add(node *decode.JsonNode) {
	d.d.Children = append(d.d.Children, node)
}

// insert 往 diffs.d.Children 中下标 i 处插入一条记录
func (d *diffs) insert(i int, node *decode.JsonNode) {
	if i > d.size() {
		d.add(node)
		return
	}
	n := make([]*decode.JsonNode, d.size()+1)
	for idx := 0; idx < i; idx++ {
		n[idx] = d.d.Children[idx]
	}
	n[i] = node
	for idx := i; idx < d.size(); idx++ {
		n[idx+1] = d.d.Children[idx]
	}
	d.d.Children = n
}

// set 将 diffs.d.Children 下标 i 处的记录修改为 node
func (d *diffs) set(i int, node *decode.JsonNode) {
	if i < len(d.d.Children) {
		d.d.Children[i] = node
	}
}

// size 返回 diffs.d.Children 切片的长度
func (d *diffs) size() int {
	return len(d.d.Children)
}

// rangeType 遍历 diffs.d.Children 并使用每一条记录的下标，值，类型作为参数执行 f
// 直到 f 返回 true 结束
func (d *diffs) rangeType(f func(i int, v *decode.JsonNode, t DiffType) bool) {
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

// ranger 遍历 diffs.d.Children 并使用每一条记录的下标，值作为参数执行 f
// // 直到 f 返回 true 结束
func (d *diffs) ranger(f func(i int, v *decode.JsonNode) bool) {
	for i, child := range d.d.Children {
		if f(i, child) {
			break
		}
	}
}

func newDiffs() *diffs {
	return &diffs{d: &decode.JsonNode{
		Type: decode.JsonNodeTypeSlice,
	}}
}
