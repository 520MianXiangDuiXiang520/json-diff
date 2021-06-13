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

package decode

import (
	"bytes"
	"github.com/valyala/bytebufferpool"
	"strings"
	"testing"
)

// 测试多种[]byte拼接方式的优劣
func BenchmarkWrite(b *testing.B) {
	length := 1000000000
	b.Run("bytes.buffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := bytes.NewBuffer([]byte{})
			for j := 0; j < length; j++ {
				buffer.WriteByte(byte('a' + j))
			}
			_ = buffer.Bytes()
		}
	})
	b.Run("strings.builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			build := strings.Builder{}
			for j := 0; j < length; j++ {
				build.WriteByte(byte('a' + j))
			}
			_ = []byte(build.String())
		}
	})
	b.Run("decode.builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			build := builder{}
			for j := 0; j < length; j++ {
				_ = build.WriteByte(byte('a' + j))
			}
			_ = build.Bytes()
		}
	})
	b.Run("[]byte", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db := make([]byte, 0)
			for j := 0; j < length; j++ {
				db = append(db, byte('a'+1))
			}
			_ = db
		}
	})

	b.Run("buffer pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pool := bytebufferpool.Get()
			for j := 0; j < length; j++ {
				_ = pool.WriteByte(byte('a' + j))
			}
			_ = pool.Bytes()
		}
	})
}
