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
	"unsafe"
)

type builder struct {
	data []byte
	// strings.Builder
}

func (b *builder) Write(d []byte) {
	b.data = append(b.data, d...)
}

func (b *builder) WriteByte(d byte) error {
	b.data = append(b.data, d)
	return nil
}

func (b *builder) Bytes() []byte {
	return b.data
}

func (b *builder) String() string {
	return *(*string)(unsafe.Pointer(&b.data))
}

// func (b *builder) getData() []byte {
//     ptr := unsafe.Pointer(b)
//     addr := (*strings.Builder)(ptr)
//     offset := unsafe.Sizeof(addr)
//     return *(*[]byte)(unsafe.Pointer(uintptr(ptr) + offset))
// }
