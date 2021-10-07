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

package json_node

import (
	"bytes"
	error2 "github.com/520MianXiangDuiXiang520/json-diff/internal/error"
	"strconv"
	"strings"
)

type ArrayNode struct {
	children []IJsonNode
	hashCode uint64
}

func NewArrayNode() *ArrayNode {
	return &ArrayNode{children: make([]IJsonNode, 0)}
}

func (a *ArrayNode) String() string {
	size := len(a.children)
	buf := bytes.NewBufferString("[")
	for i, child := range a.children {
		buf.WriteString(child.String())
		if i < size-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune(']')
	return buf.String()
}

func (a *ArrayNode) Equal(b IJsonNode) bool {
	if b.Type() != Array {
		return false
	}
	bA := b.(*ArrayNode)
	if len(a.children) != len(bA.children) {
		return false
	}
	for i, child := range a.children {
		if !child.Equal(bA.children[i]) {
			return false
		}
	}
	return true
}

func (a *ArrayNode) Hash() uint64 {
	if a.hashCode != 0 {
		return a.hashCode
	}
	code := uint64(0)
	for i, child := range a.children {
		code += (child.Hash() + (uint64(i) << (uint64(i) & 0x0000000F))) & 0x7FFFFFFF
	}
	a.hashCode = code
	return code
}

func (a *ArrayNode) reHash() {
	a.hashCode = 0
	a.Hash()
}

func (a *ArrayNode) Type() NodeType {
	return Array
}

// Add 向 Array 中根据 path 添加一个元素，path 以 / 开头，以一个
// 数字下标结束，如果当前下标处有元素，则从该下标开始所有元素向后移动
// 如果下标大小超过当前 Array 大小，中间部分会以 Null 填充. 下标以 0 开始
// Add 可能改变 array 原来的元素顺序，新元素插入后，会重新计算所有元素的哈希值
// 如果只是向末尾插入元素，请使用 Push 方法。
func (a *ArrayNode) Add(path string, v IJsonNode) error {
	if len(path) < 2 || path[0] != '/' {
		return error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return error2.PathNotFind(path)
	}
	idxS := paths[1]
	// fmt.Println(paths)
	idx, err := strconv.Atoi(idxS)
	if err != nil || idx < 0 {
		return error2.PathNotFind(path)
	}
	oldSize := len(a.children)
	if idx >= oldSize {
		for i := 0; i < idx-oldSize; i++ {
			// TODO: 使用 Mode 去配置
			a.children = append(a.children, NewNullNode())
		}
		a.children = append(a.children, v)
		a.reHash()
		return nil
	}

	old := a.children[idx]
	if old.Type() == Array || old.Type() == Object {
		newPath := strings.Join(paths[1:], "/")
		err := old.Add("/"+newPath, v)
		if err != nil {
			return err
		}
		a.reHash()
		return nil
	}

	if idx < oldSize {
		a.children = append(a.children, NewNullNode())
		for i := oldSize - 1; i >= idx; i-- {
			a.children[i+1] = a.children[i]
		}
		a.children[idx] = v
		a.reHash()
		return nil
	}

	return nil
}

// Push 向 array 的尾部插入一个元素，他可以避免 array 全部 reHash
// 因此效率比 Add 高，如果插入的元素在末尾，请尽量使用此方法
func (a *ArrayNode) Push(v IJsonNode) error {
	if v == nil {
		return error2.BadNodeType("node is nil")
	}
	a.children = append(a.children, v)
	i := len(a.children) - 1
	a.hashCode += (v.Hash() + (uint64(i) << (uint64(i) & 0x0000000F))) & 0x7FFFFFFF
	return nil
}

func (a *ArrayNode) del(path string, safe bool, v IJsonNode) (IJsonNode, bool, error) {
	if len(path) < 2 || path[0] != '/' {
		return nil, false, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, false, error2.PathNotFind(path)
	}
	idxS := paths[1]
	idx, err := strconv.Atoi(idxS)
	if err != nil || idx < 0 {
		return nil, false, error2.PathNotFind(path)
	}
	oldSize := len(a.children)
	if idx >= oldSize {
		return nil, false, error2.PathNotFind(path)
	}
	old := a.children[idx]
	if safe && !old.Equal(v) {
		return nil, false, nil
	}
	if old.Type() == Array || old.Type() == Object {
		newPath := strings.Join(paths[1:], "/")
		o, e := old.Del("/" + newPath)
		if e != nil {
			return nil, false, e
		}
		return o, true, nil
	}

	if idx < oldSize {
		for i := idx + 1; i < oldSize; i++ {
			a.children[i-1] = a.children[i]
		}
		a.children = a.children[:oldSize-1]
	}
	return old, true, nil
}

func (a *ArrayNode) Del(path string) (v IJsonNode, err error) {
	old, _, err := a.del(path, false, nil)
	if err != nil {
		return nil, err
	}
	a.reHash()
	return old, nil
}

func (a *ArrayNode) DelSafe(path string, v IJsonNode) (IJsonNode, bool, error) {
	o, ok, err := a.del(path, true, v)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	a.reHash()
	return o, true, nil
}

func (a *ArrayNode) Find(path string) IJsonNode {
	if len(path) < 2 || path[0] != '/' {
		return nil
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil
	}
	idxS := paths[1]
	idx, err := strconv.Atoi(idxS)
	if err != nil || idx < 0 {
		return nil
	}
	oldSize := len(a.children)
	if idx >= oldSize {
		return nil
	}
	old := a.children[idx]

	if old.Type() == Array || old.Type() == Object {
		newPath := strings.Join(paths[1:], "/")
		return old.Find("/" + newPath)
	}
	return old
}

func (a *ArrayNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	var err error
	defer func() {
		if err == nil {
			a.reHash()
		}
	}()
	if len(path) < 2 || path[0] != '/' {
		return nil, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, error2.PathNotFind(path)
	}
	idxS := paths[1]
	idx, err := strconv.Atoi(idxS)
	if err != nil || idx < 0 {
		return nil, error2.PathNotFind(path)
	}
	oldSize := len(a.children)
	if idx >= oldSize {
		return nil, error2.PathNotFind(path)
	}
	old := a.children[idx]

	if old.Type() == Array || old.Type() == Object {
		newPath := strings.Join(paths[1:], "/")
		return old.Replace("/"+newPath, newV)
	}

	a.children[idx] = newV

	return old, nil
}

func (a *ArrayNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	var err error
	defer func() {
		if err == nil {
			a.reHash()
		}
	}()
	if len(path) < 2 || path[0] != '/' {
		return nil, false, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, false, error2.PathNotFind(path)
	}
	idxS := paths[1]
	idx, err := strconv.Atoi(idxS)
	if err != nil || idx < 0 {
		return nil, false, error2.PathNotFind(path)
	}
	oldSize := len(a.children)
	if idx >= oldSize {
		return nil, false, error2.PathNotFind(path)
	}
	old := a.children[idx]
	if !old.Equal(oldV) {
		return nil, false, nil
	}
	if old.Type() == Array || old.Type() == Object {
		newPath := strings.Join(paths[1:], "/")
		return old.ReplaceCAS("/"+newPath, oldV, newV)
	}

	a.children[idx] = newV

	return old, true, nil
}
