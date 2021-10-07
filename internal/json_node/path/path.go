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

// Package path escape the path
// 一个正常的 path 字符串类似于: "/foo/bar/2"
// 当一个类型为 Object 的 jsonNode 的 name 中包含 "/" 时
// 可以使用 "/" 进行转义,如要获取 1 时,可以使用 path = "/foo//"
// {
//   "foo": {"/": 1}
// }
// "//" 会被转义为 ~1, 而原来 name 中包含的 ~1 会被转义为 ~01
// ~01 会被转义为 ~001 依次类推
package path

import (
	"bytes"
	"regexp"
	"strings"
)

type IPath interface {
	Next() string
}

type Path struct {
	p string
}

func (p *Path) Next() string {
	return ""
}

var keyReplaceRegexp = regexp.MustCompile(`~0*1`)

// KeyReplace 转义 key 中的特殊字符
// "/"    会被替换成 "~1"
// "~1"   会被替换成 "~01"
// "~01"  会被替换为 "~001"
// "~001" 会被替换为 "~0001"
// 依此类推
func KeyReplace(key string) string {
	resList := keyReplaceRegexp.FindAllStringIndex(key, -1)
	buff := bytes.NewBufferString("")
	pre := 0
	for _, res := range resList {
		buff.WriteString(key[pre:res[0]])
		buff.WriteRune('~')
		for i := 1; i < res[1]-res[0]; i++ {
			buff.WriteRune('0')
		}
		buff.WriteRune('1')
		pre = res[1]
	}
	buff.WriteString(key[pre:])
	return strings.ReplaceAll(buff.String(), "/", "~1")
}

func KeyRestore(key string) string {
	key = strings.ReplaceAll(key, "~1", "/")
	resList := keyReplaceRegexp.FindAllStringIndex(key, -1)
	buff := bytes.NewBufferString("")
	pre := 0
	for _, res := range resList {
		buff.WriteString(key[pre:res[0]])
		buff.WriteRune('~')
		for i := 3; i < res[1]-res[0]; i++ {
			buff.WriteRune('0')
		}
		buff.WriteRune('1')
		pre = res[1]
	}
	buff.WriteString(key[pre:])
	return buff.String()
}

func NewPath(p string) *Path {
	return nil
}
