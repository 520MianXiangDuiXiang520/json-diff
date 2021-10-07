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

// Package parse 实现了一个基于 RFC 4627 的 JSON 解析器
// 它提供以下接口用于 JSON 字符串，JsonNode 对象和 go interface 之间的相互转换。
// Marshal: 用于将 json_node.IJsonNode 对象序列化为 byte 数组（[]byte） :
//
//   node := json_node.NewObjectNode()
//   res := Marshal(node)
//   fmt.Println(string(res)
//
// Unmarshal: 用于将一个JSON 串转换为 json_node.IJsonNode 对象：
// 输入的 Json 串不满足 RFC 4627 标准时，返回 UnmarshalErr
//
//   str := `{"foo": ["bar"]}`
//   node, err := Unmarshal([]byte(str))
//   if err != nil {
//       fmt.Println(err)
//   }
//   node.Add("/bar", json_node.NewObjectNode())
//
// Build: 用于将一个 map[string]interface{} 或 struct{}
// 转化为 json_node.IJsonNode 对象，struct 转换规则与标签与
// 官方 json 库保持一致。输入为 nil 或转换失败时，返回 BuildErr
//
//   var obj = struct {
//      Foo       []string `json:"foo"`
//      Invisible string   `json:"-"`
//   }{
//      Foo:       []string{"bar"},
//      Invisible: "Invisible",
//   }
//   node, err := Build(obj)
//   if err != nil {
//       fmt.Println(err)
//   }
//   node.Add("/bar", json_node.NewObjectNode())
package parse

import (
	"github.com/520MianXiangDuiXiang520/json-diff/internal/json_node"
)

func Marshal(node json_node.IJsonNode) []byte {
	return nil
}

func Unmarshal(data []byte) (json_node.IJsonNode, error) {
	parse := NewParser()
	node, err := parse.Parser(data)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func Build(v interface{}) (json_node.IJsonNode, error) {
	return nil, nil
}
