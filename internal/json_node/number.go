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

type NumberNode struct {
	StringNode
}

func NewNumberNode(value []byte) *NumberNode {
	n := &NumberNode{}
	n.value = value
	n.hashCode = pkg.SDBMHash(value)
	return n
}

func (n *NumberNode) String() string {
	return string(n.value)
}

func (n *NumberNode) Type() NodeType {
	return Number
}

func (n *NumberNode) Find(path string) IJsonNode {
	if path == "/" {
		return n
	}
	return nil
}

func (n *NumberNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	if path == "/" {
		if newV.Type() != Number {
			return nil, error2.BadNodeType("only the number type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		v := newV.(*NumberNode)
		old := &NumberNode{}
		old.hashCode = n.hashCode
		old.value = n.value

		n.value = v.value
		n.hashCode = v.hashCode
		return old, nil
	}
	return nil, error2.PathNotFind(path)
}

func (n *NumberNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	if path == "/" {
		if newV.Type() != Number {
			return nil, false, error2.BadNodeType("only the number type node is" +
				" allowed to be replaced, you give a " + TypeNames[newV.Type()])
		}
		if !n.Equal(oldV) {
			return nil, false, nil
		}
		v := newV.(*NumberNode)
		old := &NumberNode{}
		old.hashCode = n.hashCode
		old.value = n.value

		n.value = v.value
		n.hashCode = v.hashCode
		return old, true, nil
	}
	return nil, false, error2.PathNotFind(path)
}
