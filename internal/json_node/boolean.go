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
	error2 "github.com/520MianXiangDuiXiang520/json-diff/internal/error"
)

type BooleanNode struct {
	value bool
}

func NewBooleanNode(v bool) *BooleanNode {
	return &BooleanNode{value: v}
}

func (b *BooleanNode) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b *BooleanNode) Equal(n IJsonNode) bool {
	if n.Type() != Boolean {
		return false
	}
	return n.(*BooleanNode).value == b.value
}

func (b *BooleanNode) Hash() uint64 {
	if b.value {
		return 0x0f0f0f0f
	}
	return 0xf0f0f0f0
}

func (b *BooleanNode) Type() NodeType {
	return Boolean
}

func (b *BooleanNode) Add(path string, v IJsonNode) error {
	return error2.BadNodeType("boolean type cannot add child node")
}

func (b *BooleanNode) Del(path string) (v IJsonNode, err error) {
	return nil, error2.BadNodeType("boolean type cannot do del")
}

func (b *BooleanNode) DelSafe(path string, v IJsonNode) (IJsonNode, bool, error) {
	return nil, false, error2.BadNodeType("boolean type cannot do delSafe")
}

func (b *BooleanNode) Find(path string) IJsonNode {
	if path == "/" {
		return b
	}
	return nil
}

func (b *BooleanNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	if path == "/" {
		if newV.Type() != Boolean {
			return nil, error2.BadNodeType("only the boolean type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		old := &BooleanNode{value: b.value}
		b.value = newV.(*BooleanNode).value
		return old, nil
	}
	return nil, error2.PathNotFind(path)
}

func (b *BooleanNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	if path == "/" {
		if newV.Type() != Boolean {
			return nil, false, error2.BadNodeType("only the boolean type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		if !b.Equal(oldV) {
			return nil, false, nil
		}
		old := &BooleanNode{value: b.value}
		b.value = newV.(*BooleanNode).value
		return old, true, nil
	}
	return nil, false, error2.PathNotFind(path)
}
