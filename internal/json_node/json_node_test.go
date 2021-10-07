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
	"fmt"
	"testing"
)

func TestNewObjectNode(t *testing.T) {
	root := NewObjectNode()
	age := NewNumberNode([]byte("18"))
	score := NewArrayNode()
	err := score.Add("/0", NewStringNode([]byte("110")))
	if err != nil {
		t.Error(err)
		return
	}
	err = score.Add("/1", NewNumberNode([]byte("120")))
	if err != nil {
		t.Error()
		return
	}
	err = score.Add("/3", NewBooleanNode(false))
	if err != nil {
		t.Error()
		return
	}
	err = score.Add("/5", NewArrayNode())
	if err != nil {
		t.Error()
		return
	}
	err = score.Add("/7", NewObjectNode())
	if err != nil {
		t.Error()
		return
	}
	deleted, err := score.Del("/6")
	if err != nil {
		t.Error()
		return
	}
	if deleted.Type() != Null {
		t.Error(deleted)
	}
	err = root.Add("/age", age)
	if err != nil {
		t.Error()
		return
	}
	err = root.Add("/score", score)
	if err != nil {
		t.Error()
		return
	}
	fmt.Println(root)
}
