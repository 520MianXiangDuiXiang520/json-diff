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
	"fmt"
	"github.com/pkg/errors"
)

func (jn *JsonNode) marshalValue() (*builder, error) {
	tokens := &builder{}
	if jn.Value == nil {
		tokens.Write(tokenToBytes(NULL, nil, nil))
		return tokens, nil
	}
	switch jn.Value.(type) {
	case string:
		tokens.Write(tokenToBytes(STRING, jn.Value, jn.originalValue))
	case int, int8, int16, int32, int64, float64,
		float32, uint, uint8, uint16, uint32, uint64:
		tokens.Write(tokenToBytes(NUMBER, jn.Value, jn.originalValue))
	case bool:
		tokens.Write(tokenToBytes(Boolean, jn.Value, nil))
	default:
		return nil, errors.New(fmt.Sprintf("fail to marshal node: %v", jn.Value))
	}
	return tokens, nil
}

func (jn *JsonNode) marshalArray() (*builder, error) {
	tokens := &builder{}
	size := len(jn.Children)
	for i, child := range jn.Children {
		childToken, err := child.marshal()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tokens.Write(childToken.Bytes())
		if i+1 < size {
			tokens.Write(tokenToBytes(Comma, nil, nil))
		}
	}
	return tokens, nil
}

func (jn *JsonNode) marshalObject() (*builder, error) {
	tokens := &builder{}
	size := len(jn.ChildrenMap)
	index := 0
	for key, value := range jn.ChildrenMap {
		childTokens, err := value.marshalPair(key)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tokens.Write(childTokens.Bytes())
		if index+1 < size {
			tokens.Write(tokenToBytes(Comma, nil, nil))
		}
		index++
	}
	return tokens, nil
}

func (jn *JsonNode) marshalPair(key string) (*builder, error) {
	tokens := &builder{}
	value, err := jn.marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tokens.Write(tokenToBytes(STRING, key, nil))
	tokens.Write(tokenToBytes(Colon, nil, nil))
	tokens.Write(value.Bytes())
	return tokens, nil
}

func (jn *JsonNode) marshal() (*builder, error) {
	tokens := &builder{}
	switch jn.Type {
	case JsonNodeTypeValue:
		vTokens, err := jn.marshalValue()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tokens.Write(vTokens.Bytes())
	case JsonNodeTypeSlice:
		tokens.Write(tokenToBytes(StartArray, nil, nil))
		aTokens, err := jn.marshalArray()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tokens.Write(aTokens.Bytes())
		tokens.Write(tokenToBytes(EndArray, nil, nil))
	case JsonNodeTypeObject:
		tokens.Write(tokenToBytes(StartObj, nil, nil))
		oTokens, err := jn.marshalObject()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		tokens.Write(oTokens.Bytes())
		tokens.Write(tokenToBytes(EndObj, nil, nil))
	}
	return tokens, nil
}

// Marshal 将一个 JsonNode 对象序列化为 Json 字符。
func (jn *JsonNode) Marshal() ([]byte, error) {
	tokens, err := jn.marshal()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tokens.Write(tokenToBytes(EndDoc, nil, nil))
	return tokens.Bytes(), nil
}

// Marshal 将一个 JsonNode 对象序列化为 Json 字符。
func Marshal(node *JsonNode) ([]byte, error) {
	return node.Marshal()
}
