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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

func TestUnmarshal(t *testing.T) {
	var builder strings.Builder
	builder.WriteString("xw")
	ptr := unsafe.Pointer(&builder)
	addr := (*strings.Builder)(ptr)
	offset := unsafe.Sizeof(addr)
	buf := *(*[]byte)(unsafe.Pointer(uintptr(ptr) + offset))
	fmt.Println(buf, builder.String())
}

// 先反序列化成 interface{}, 再序列化成 []byte, 再反序列化到
func TestMarshal(t *testing.T) {
	args := []struct {
		name  string
		input string
	}{
		{"empty object", "{}"},
		{"empty array", "[]"},
		{"only number zero", "0"},
		{"only number E", "0.4E-32"},
		{"only number", "2021"},
		{"only number2", "-0.8"},
		{"only number3", "-0.8e23"},
		{"only string", `"aaaa"`},
		{"only false", "false"},
		{"only true", "true"},
		{"only null", "null"},
		{"only null", `{"a":[1.2]}`},
		{"object", `{"a": 1, "b": "123", "c": false, "d": null}`},
		{"array", `[1, "2", false, null, [1, 2.5, {}], {"a": 1, "b": null}]`},
		{"complex", `{"a": null, "b": false, "c": "奤","d": [
  1, 2, null, "d", [false, true, null, {}, [], {
    "a": null, "b": false, "c": [], "d": {
      "a": false, "b": {
        "a": null, "b": -0.9E2, "c": true, "e": [], "f": "XXX", "g": "时候就是"
      }
    }
  }]
], "e": {}, "f": []}`},
	}
	for _, arg := range args {
		t.Run(arg.name, func(st *testing.T) {
			input := arg.input
			var inf interface{}
			err := json.Unmarshal([]byte(input), &inf)
			if err != nil {
				st.Errorf("got an error %+v", err)
			}
			node, err := Unmarshal([]byte(input))
			if err != nil {
				st.Errorf("got an error %+v", err)
			}
			afterMarshal, err := Marshal(node)
			if err != nil {
				st.Errorf("got an error %+v", err)
			}
			var infAfterMarshal interface{}
			err = json.Unmarshal(afterMarshal, &infAfterMarshal)
			if err != nil {
				st.Errorf("got an error %+v", err)
			}
			if !compareInterface(inf, infAfterMarshal) {
				st.Errorf("not equal, want %v, got %v", inf, infAfterMarshal)
			}
		})
	}
}

func compareInterface(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	switch a.(type) {
	case map[string]interface{}:
		switch b.(type) {
		case map[string]interface{}:
			av, bv := a.(map[string]interface{}), b.(map[string]interface{})
			if len(av) != len(bv) {
				return false
			}
			for k, v := range av {
				if !compareInterface(v, bv[k]) {
					return false
				}
			}
		default:
			return false
		}
	case []interface{}:
		switch b.(type) {
		default:
			return false
		case []interface{}:
			av, bv := a.([]interface{}), b.([]interface{})
			if len(av) != len(bv) {
				return false
			}
			for i, v := range av {
				if !compareInterface(v, bv[i]) {
					return false
				}
			}
		}
	case int, int8, int16, int32, int64, float64, float32, string, bool:
		return a == b
	default:
		return false
	}
	return true
}
