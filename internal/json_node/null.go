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

type NullNode struct{}

func NewNullNode() *NullNode {
	return &NullNode{}
}

func (n *NullNode) String() string {
	return "null"
}

func (n *NullNode) Equal(b IJsonNode) bool {
	return b.Type() == n.Type()
}

func (n *NullNode) Hash() uint64 {
	return 0xffffffff
}

func (n *NullNode) Type() NodeType {
	return Null
}

func (n *NullNode) Add(path string, v IJsonNode) error {
	return error2.BadNodeType(TypeNames[n.Type()] + " type cannot add child node")
}

func (n *NullNode) Del(path string) (v IJsonNode, err error) {
	return nil, error2.BadNodeType(TypeNames[n.Type()] + " type cannot do del")
}

func (n *NullNode) DelSafe(path string, v IJsonNode) (IJsonNode, bool, error) {
	return nil, false, error2.BadNodeType(TypeNames[n.Type()] + " type cannot do delSafe")
}

func (n *NullNode) Find(path string) IJsonNode {
	if path == "/" {
		return n
	}
	return nil
}

// Replace 由于 null 的值不可变，所以对 null 节点执行 Replace 是无效的
func (n *NullNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	return nil, nil
}

func (n *NullNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	return nil, false, nil
}
