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

type NodeType int8

const (
	Object = iota + 1
	Array
	String
	Number
	Null
	Boolean
)

var (
	TypeNames = map[NodeType]string{
		Object:  "object",
		Array:   "array",
		String:  "string",
		Number:  "number",
		Null:    "null",
		Boolean: "boolean",
	}
)

// IJsonNode 表示一个 Json 对象
type IJsonNode interface {
	String() string

	// Equal 比较当前节点与给定的 b 是否相等。
	Equal(b IJsonNode) bool

	// Hash 返回当前节点的一个 64 位哈希值
	Hash() uint64

	// Type 返回当前节点的类型
	Type() NodeType

	// Add 往当前节点的固定路径插入一个子节点，如果路径不存在
	// Add 将会返回一个 error.PathNotFind 错误，如果当前节点不允许
	// 设置子节点，Add 将返回一个 error.NodeTypeError
	Add(path string, v IJsonNode) error

	// Del 删除当前节点固定路径的值并返回被删除了的值
	// 如果 path 不存在，将会返回一个 error.PathNotFind 错误
	// Del 不会检查 path 对应的值是否正确，想要使用此功能，请使用
	// DelSafe 方法
	Del(path string) (v IJsonNode, err error)

	// DelSafe 类似于 Del 只不过在删除前他会对比 path 对应的节点是否与 v 相等
	// 如果不相等，DelSafe 会放弃删除，第二个返回值用于表示是否进行了删除
	DelSafe(path string, v IJsonNode) (IJsonNode, bool, error)

	// Find 返回 path 对应的值，如果 path 不存在，返回 nil
	Find(path string) IJsonNode

	// Replace 使用 newV 替换 path 对应的值并返回旧的值，如果 path 不存在
	// Replace 将返回 error.PathNotFind 错误
	Replace(path string, newV IJsonNode) (IJsonNode, error)

	// ReplaceCAS 类似于 Replace, 只不过在替换前会比较 path
	// 对应的值是否与 oldV 相等，只有相等时才会进行替换,
	// 第二个返回值用于表示是否进行了替换
	ReplaceCAS(path string, oldV, newV IJsonNode) (IJsonNode, bool, error)
}
