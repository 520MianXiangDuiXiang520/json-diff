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
	"github.com/520MianXiangDuiXiang520/json-diff/pkg"
	"strings"
)

type ObjectNode struct {
	children map[string]IJsonNode
	hashCode uint64
}

func NewObjectNode() *ObjectNode {
	return &ObjectNode{
		children: make(map[string]IJsonNode),
	}
}

func (o *ObjectNode) String() string {
	// TODO: / 转义
	size := len(o.children)
	buf := bytes.NewBufferString("{")
	idx := 0
	for key, child := range o.children {
		buf.WriteRune('"')
		buf.WriteString(key)
		buf.WriteRune('"')
		buf.WriteRune(':')
		buf.WriteString(child.String())
		if idx < size-1 {
			buf.WriteRune(',')
		}
		idx++
	}
	buf.WriteRune('}')
	return buf.String()
}

func (o *ObjectNode) Equal(b IJsonNode) bool {
	if o.Hash() != b.Hash() {
		return false
	}
	return o.slowEqual(b)
}

func (o *ObjectNode) slowEqual(b IJsonNode) bool {
	if b.Type() != Object {
		return false
	}
	ob := b.(*ObjectNode)
	for name, node := range o.children {
		nodeB, ok := ob.children[name]
		if !ok {
			return false
		}
		if !nodeB.Equal(node) {
			return false
		}
	}
	return true
}

func (o *ObjectNode) Hash() uint64 {
	if o.hashCode != 0 {
		return o.hashCode
	}
	code := uint64(0)
	for name, node := range o.children {
		// fmt.Println(name, node)
		code += (pkg.SDBMHash([]byte(name)) + node.Hash()) & 0x7FFFFFFF
	}
	o.hashCode = code
	return code
}

func (o *ObjectNode) reHash() {
	o.hashCode = 0
	o.Hash()
}

func (o *ObjectNode) Type() NodeType {
	return Object
}

func (o *ObjectNode) Add(path string, v IJsonNode) error {
	if len(path) < 2 || path[0] != '/' {
		return error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return error2.PathNotFind(path)
	}
	startKey := paths[1]
	got, ok := o.children[startKey]
	if ok {
		if got.Type() == Array || got.Type() == Object {
			newPath := "/" + strings.Join(paths[1:], "/")
			err := got.Add(newPath, v)
			if err != nil {
				return err
			}
			return nil
		}
		return error2.KeyExistedF(path)
	}
	o.children[startKey] = v
	o.hashCode += (pkg.SDBMHash([]byte(startKey)) + v.Hash()) & 0x7FFFFFFF
	return nil
}

func (o *ObjectNode) Del(path string) (v IJsonNode, err error) {
	if len(path) < 2 || path[0] != '/' {
		return nil, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, error2.PathNotFind(path)
	}
	startKey := paths[1]
	got, ok := o.children[startKey]
	if !ok {
		return nil, error2.PathNotFind(path)
	}
	if got.Type() == Array || got.Type() == Object {

		newPath := "/" + strings.Join(paths[1:], "/")
		old, err := got.Del(newPath)
		if err != nil {
			return nil, err
		}
		o.reHash()
		return old, nil
	}
	delete(o.children, startKey)
	o.reHash()
	return v, nil
}

func (o *ObjectNode) DelSafe(path string, old IJsonNode) (IJsonNode, bool, error) {
	if len(path) < 2 || path[0] != '/' {
		return nil, false, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, false, error2.PathNotFind(path)
	}
	startKey := paths[1]
	v, ok := o.children[startKey]
	if !ok {
		return nil, false, error2.PathNotFind(path)
	}
	if !v.Equal(old) {
		return nil, false, error2.UnEqualF(old.String(), v.String())
	}
	if v.Type() == Array || v.Type() == Object {

		newPath := "/" + strings.Join(paths[1:], "/")
		ol, deleted, err := v.DelSafe(newPath, v)
		if err != nil {
			return nil, false, err
		}
		if !deleted {
			return nil, false, nil
		}
		o.reHash()
		return ol, true, nil
	}
	delete(o.children, startKey)
	o.reHash()
	return v, true, nil
}

func (o *ObjectNode) Find(path string) IJsonNode {
	if len(path) < 1 || path[0] != '/' {
		return nil
	}
	if path == "/" {
		return o
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil
	}
	startKey := paths[1]
	v, ok := o.children[startKey]
	if !ok {
		return nil
	}
	if v.Type() == Array || v.Type() == Object {
		newPath := "/" + strings.Join(paths[1:], "/")
		return v.Find(newPath)

	}
	return v
}

func (o *ObjectNode) Replace(path string, newV IJsonNode) (IJsonNode, error) {
	if len(path) < 2 || path[0] != '/' {
		return nil, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, error2.PathNotFind(path)
	}
	startKey := paths[1]
	v, ok := o.children[startKey]
	if !ok {
		return nil, error2.PathNotFind(path)
	}
	if v.Type() == Array || v.Type() == Object {

		newPath := "/" + strings.Join(paths[1:], "/")
		old, err := v.Replace(newPath, newV)
		if err != nil {
			return nil, err
		}
		o.reHash()
		return old, nil
	}
	o.children[startKey] = newV
	o.reHash()
	return v, nil
}

func (o *ObjectNode) ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error) {
	if len(path) < 2 || path[0] != '/' {
		return nil, false, error2.PathNotFind(path)
	}
	paths := strings.Split(path, "/")
	if len(paths) < 2 {
		return nil, false, error2.PathNotFind(path)
	}
	startKey := paths[1]
	v, ok := o.children[startKey]
	if !ok {
		return nil, false, error2.PathNotFind(path)
	}
	if !v.Equal(oldV) {
		return nil, false, error2.UnEqualF(oldV.String(), v.String())
	}
	if v.Type() == Array || v.Type() == Object {

		newPath := "/" + strings.Join(paths[1:], "/")
		old, err := v.Replace(newPath, newV)
		if err != nil {
			return nil, false, err
		}
		o.reHash()
		return old, true, nil
	}
	o.children[startKey] = newV
	o.reHash()
	return v, true, nil
}
