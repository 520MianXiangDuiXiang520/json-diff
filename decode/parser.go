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
	"github.com/pkg/errors"
)

var parserError = errors.New("fail to parser json")

// 语法分析器
// doc = object | array | string | number | true | false | null
func (l *jsonParser) parser() error {
	for {
		if l.parserOffset >= len(l.tokens) {
			break
		}
		token := l.tokens[l.parserOffset]
		if token.t == EndDoc {
			break
		}
		switch token.t {
		case StartObj:
			obj, err := l.parserObj(0)
			if err != nil {
				return err
			}
			l.jsonNode = obj
		case StartArray:
			arr, err := l.parserArray(0)
			if err != nil {
				return err
			}
			l.jsonNode = arr
		case Boolean, NULL:
			node := NewValueNode(token.v, 0)
			l.parserOffset++
			l.jsonNode = node
		case NUMBER, STRING:
			node := newOriginalValueNode(token.originalValue, token.v, 0)
			l.parserOffset++
			l.jsonNode = node
		default:
			return parserError
		}
	}
	return nil
}

// array = [] | [ elements ]
func (l *jsonParser) parserArray(level int) (*JsonNode, error) {
	l.parserOffset++
	if l.parserOffset >= len(l.tokens) {
		return nil, parserError
	}
	token := l.tokens[l.parserOffset]
	switch token.t {
	case EndArray:
		l.parserOffset++
		return NewSliceNode(make([]*JsonNode, 0), level), nil
	// case NUMBER:
	//     node := newOriginalValueNode(token.originalValue, make([]*JsonNode, 0), level)
	//     err := l.parseElements(level, node)
	//     if err != nil {
	//         return nil, err
	//     }
	//     return node, nil
	case STRING, NUMBER, StartObj, StartArray, Boolean, NULL:
		node := NewSliceNode(make([]*JsonNode, 0), level)
		err := l.parseElements(level, node)
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, parserError
	}
}

// elements = value  | value , elements
func (l *jsonParser) parseElements(level int, parent *JsonNode) error {
	for {
		err := l.parseValue(level, parent)
		if err != nil {
			return err
		}
		if l.parserOffset >= len(l.tokens) {
			return parserError
		}
		first := l.tokens[l.parserOffset]
		if first.t == Comma {
			l.parserOffset++
			continue
		} else if first.t == EndArray {
			l.parserOffset++
			break
		} else {
			return parserError
		}
	}
	return nil
}

// value = string | number | object | array | true | false | null
func (l *jsonParser) parseValue(level int, parent *JsonNode) error {
	if l.parserOffset >= len(l.tokens) {
		return parserError
	}
	token := l.tokens[l.parserOffset]
	switch token.t {
	case Boolean, NULL:
		childNode := NewValueNode(token.v, level+1)
		_ = parent.Append(childNode)
		l.parserOffset++
	case NUMBER, STRING:
		childNode := newOriginalValueNode(token.originalValue, token.v, level+1)
		_ = parent.Append(childNode)
		l.parserOffset++
	case StartObj:
		childNode, err := l.parserObj(level + 1)
		if err != nil {
			return parserError
		}
		_ = parent.Append(childNode)
	case StartArray:
		childNode, err := l.parserArray(level + 1)
		if err != nil {
			return parserError
		}
		_ = parent.Append(childNode)
	default:
		return parserError
	}
	return nil
}

// object = {} | { members }
func (l *jsonParser) parserObj(level int) (*JsonNode, error) {
	l.parserOffset++
	if l.parserOffset >= len(l.tokens) {
		return nil, parserError
	}
	token := l.tokens[l.parserOffset]
	switch token.t {
	case EndObj:
		l.parserOffset++
		return NewObjectNode("", map[string]*JsonNode{}, level), nil
	case STRING:
		node := NewObjectNode("", map[string]*JsonNode{}, level)
		err := l.parseMembers(level, node)
		if err != nil {
			return nil, err
		}
		return node, nil
	default:
		return nil, parserError
	}
}

// members = pair | pair , members
func (l *jsonParser) parseMembers(level int, parent *JsonNode) error {
	for {
		err := l.parsePair(level, parent)
		if err != nil {
			return err
		}
		if l.parserOffset >= len(l.tokens) {
			return parserError
		}
		first := l.tokens[l.parserOffset]
		if first.t == Comma {
			l.parserOffset++
			continue
		} else if first.t == EndObj {
			l.parserOffset++
			break
		} else {
			return parserError
		}
	}
	return nil
}

// pair = string : value
// value = string | number | object | array | true | false | null
func (l *jsonParser) parsePair(level int, parent *JsonNode) error {
	// the parser offset pointer in string
	if l.parserOffset+2 >= len(l.tokens) {
		return parserError
	}
	idx := l.parserOffset
	first, second, third := l.tokens[idx], l.tokens[idx+1], l.tokens[idx+2]
	if first.t != STRING || second.t != Colon {
		return parserError
	}
	l.parserOffset += 2
	k := first.v.(string)
	switch third.t {
	case Boolean, NULL:
		childNode := NewValueNode(third.v, level+1)
		_ = parent.ADD(k, childNode)
		l.parserOffset++
	case NUMBER, STRING:
		childNode := newOriginalValueNode(third.originalValue, third.v, level+1)
		_ = parent.ADD(k, childNode)
		l.parserOffset++
	case StartObj:
		childNode, err := l.parserObj(level + 1)
		if err != nil {
			return parserError
		}
		_ = parent.ADD(k, childNode)
	case StartArray:
		childNode, err := l.parserArray(level + 1)
		if err != nil {
			return parserError
		}
		_ = parent.ADD(k, childNode)
	default:
		return parserError
	}
	return nil
}

// Unmarshal 将一个 json 序列格式化为 JsonNode 对象
func Unmarshal(input []byte) (*JsonNode, error) {
	if input == nil {
		return nil, nil
	}
	l := initLexer(input)
	err := l.tokenizer()
	if err != nil {
		return nil, errors.Wrap(err, "fail to Unmarshal")
	}
	err = l.parser()
	if err != nil {
		return nil, errors.Wrap(err, "fail to Unmarshal")
	}
	return l.jsonNode, nil
}
