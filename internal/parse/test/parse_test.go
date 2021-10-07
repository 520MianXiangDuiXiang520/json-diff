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

package test

import (
	"fmt"
	"github.com/520MianXiangDuiXiang520/json-diff/internal/parse"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func TestUnmarshal(t *testing.T) {
	files, _ := ioutil.ReadDir("./test_data")
	for _, f := range files {
		ns := strings.Split(f.Name(), ".")
		if len(ns) < 2 {
			continue
		}
		ty := ns[len(ns)-1]
		if strings.ToLower(ty) != "json" {
			continue
		}
		// if f.Name() != "only_string2.json" {
		//     continue
		// }
		t.Run(f.Name(), func(t *testing.T) {
			data, err := ioutil.ReadFile("./test_data/" + f.Name())
			if err != nil {
				t.Errorf("fail to read fail %s", f.Name())
				return
			}
			_, err = parse.Unmarshal(data)
			if err != nil {
				t.Errorf("fail to Unmarshal data, err: %+v ", err)
				return
			}
			// fmt.Println(node.String())
		})
	}
}

func TestUnmarshal2(t *testing.T) {
	data, err := ioutil.ReadFile("./test_data/big_string.json")
	if err != nil {
		t.Errorf("fail to read fail %s", "big.json")
		return
	}
	n := time.Now()
	f, err := os.Create(fmt.Sprintf("./pprof/%d_%d_%d_%d_%d",
		n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute()))
	if err != nil {
		fmt.Println(err)
		panic("[Debug] fail to create pprof file")
	}
	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()
	for i := 0; i < 10; i++ {
		_, err := parse.Unmarshal(data)
		if err != nil {
			t.Errorf("fail to Unmarshal data, err: %+v ", err)
			return
		}
	}
	// fmt.Println(node.String())
}
