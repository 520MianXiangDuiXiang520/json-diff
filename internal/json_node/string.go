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
	"github.com/520MianXiangDuiXiang520/json-diff/pkg"
)

type StringNode struct {
	value    []byte // 使用 string 表示 number,避免精度丢失
	hashCode uint64
}

func NewStringNode(value []byte) *StringNode {
	return &StringNode{
		value:    value,
		hashCode: pkg.SDBMHash(value),
	}
}

func (n *StringNode) String() string {
	return "\"" + string(n.value) + "\""
}

func (n *StringNode) Equal(b IJsonNode) bool {
	if b.Type() != n.Type() {
		return false
	}
	bN := b.(*StringNode)
	return string(n.value) == string(bN.value)
}

func (n *StringNode) Hash() uint64 {
	if n.hashCode != 0 {
		return n.hashCode
	}
	n.hashCode = pkg.SDBMHash(n.value)
	return n.hashCode
}

func (n *StringNode) Type() NodeType {
	return String
}

func (n *StringNode) Add(path string, v IJsonNode) error {
	return error2.BadNodeType(TypeNames[n.Type()] + " type cannot add child node")
}

func (n *StringNode) Del(path string) (v IJsonNode, err error) {
	return nil, error2.PathNotFind(TypeNames[n.Type()] + " type can not do Del")
}

func (n *StringNode) DelSafe(path string, v IJsonNode) (IJsonNode, bool, error) {
	return nil, false, error2.PathNotFind(TypeNames[n.Type()] + " type can not do DelSafe")
}

func (n *StringNode) Find(path string) IJsonNode {
	if path == "/" {
		return n
	}
	return nil
}

func (n *StringNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	if path == "/" {
		if newV.Type() != String {
			return nil, error2.BadNodeType("only the string type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		v := newV.(*StringNode)
		old := &StringNode{
			value:    n.value,
			hashCode: n.hashCode,
		}
		n.value = v.value
		n.hashCode = v.hashCode
		return old, nil
	}
	return nil, error2.PathNotFind(path)
}

func (n *StringNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	if path == "/" {
		if newV.Type() != String {
			return nil, false, error2.BadNodeType("only the string type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		if !n.Equal(oldV) {
			return nil, false, nil
		}
		v := newV.(*StringNode)
		old := &StringNode{
			value:    n.value,
			hashCode: n.hashCode,
		}
		n.value = v.value
		n.hashCode = v.hashCode
		return old, true, nil
	}
	return nil, false, error2.PathNotFind(path)
}
