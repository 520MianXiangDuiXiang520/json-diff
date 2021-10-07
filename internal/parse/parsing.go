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

package parse

import (
	"github.com/520MianXiangDuiXiang520/json-diff/internal/json_node"
	"github.com/pkg/errors"
)

// Parsing 语法分析器
type Parsing struct {
	parser *JsonParser
	node   json_node.IJsonNode
	idx    int
}

func NewParsing(p *JsonParser) *Parsing {
	return &Parsing{parser: p}
}

func (p *Parsing) Scanner() error {
	firstToken, err := p.getNext()
	if err != nil {
		return err
	}
	return p.handler(firstToken)
}

func (p *Parsing) getNext() (*Token, error) {
	if p.idx >= len(p.parser.tokenList) {
		return nil, errors.WithMessagef(ParsingError, "no token can be used")
	}
	t := p.parser.tokenList[p.idx]
	p.idx++
	return t, nil
}

func (p *Parsing) handler(t *Token) error {
	switch t.Type() {
	case StartObj:
		obj, err := p.parsingObj()
		if err != nil {
			return err
		}
		p.node = obj
	case StartArray:
		arr := json_node.NewArrayNode()
		err := p.parsingArray(arr)
		if err != nil {
			return err
		}
		p.node = arr
	case STRING, NULL, True, False, NUMBER:
		endDocToken, err := p.getNext()
		if err != nil {
			return err
		}
		if endDocToken.Type() != EndDoc {
			return errors.WithMessagef(ParsingError, "token: %s", endDocToken.String())
		}
		p.node = getNodeBySimpleToken(t)
	case EndDoc:
		return nil
	default:
		return errors.WithMessagef(ParsingError, "token: %s", t.String())
	}
	return nil
}

func (p *Parsing) parsingObj() (json_node.IJsonNode, error) {
	token, err := p.getNext()
	if err != nil {
		return nil, err
	}
	root := json_node.NewObjectNode()
	switch token.Type() {
	case EndObj:
		return root, nil
	case STRING:
		err := p.pair(root, token.String())
		if err != nil {
			return nil, err
		}
		return root, nil
	default:
		return nil, errors.WithMessagef(ParsingError, "token: %s", token.String())
	}
}

func (p *Parsing) pair(root *json_node.ObjectNode, key string) error {
	wantColon, err := p.getNext()
	if err != nil {
		return err
	}
	if wantColon.Type() != Colon {
		return errors.WithMessagef(ParsingError, "want a colon(,) got a %s", wantColon.String())
	}
	valueToken, err := p.getNext()
	if err != nil {
		return err
	}
	switch valueToken.Type() {
	case True, False, NULL, STRING, NUMBER:
		err := root.Add("/"+key, getNodeBySimpleToken(valueToken))
		if err != nil {
			return err
		}
	case StartObj:
		value, err := p.parsingObj()
		if err != nil {
			return err
		}
		err = root.Add("/"+key, value)
		if err != nil {
			return err
		}
	case StartArray:
		value := json_node.NewArrayNode()
		err := p.parsingArray(value)
		if err != nil {
			return err
		}
		err = root.Add("/"+key, value)
		if err != nil {
			return err
		}
	default:
		return errors.WithMessagef(ParsingError, "token: %s", valueToken.String())
	}

	endToken, err := p.getNext()
	if err != nil {
		return err
	}
	switch endToken.Type() {
	case EndObj:
		return nil
	case Comma:
		keyToken, err := p.getNext()
		if err != nil {
			return err
		}
		if keyToken.Type() != STRING {
			return errors.WithMessagef(ParsingError, "token: %s", keyToken)
		}
		err = p.pair(root, keyToken.String())
		if err != nil {
			return err
		}
	default:
		return errors.WithMessagef(ParsingError, "token: %s", endToken)
	}

	return nil
}

func (p *Parsing) parsingArray(root *json_node.ArrayNode) error {
	token, err := p.getNext()
	if err != nil {
		return err
	}

	switch token.Type() {
	case EndArray:
		return nil
	case STRING, True, False, NULL, NUMBER:
		err := root.Push(getNodeBySimpleToken(token))
		if err != nil {
			return err
		}
	case StartArray:
		c := json_node.NewArrayNode()
		err := p.parsingArray(c)
		if err != nil {
			return err
		}
		err = root.Push(c)
		if err != nil {
			return err
		}
	case StartObj:
		v, err := p.parsingObj()
		if err != nil {
			return err
		}
		err = root.Push(v)
		if err != nil {
			return err
		}
	default:
		return errors.WithMessagef(ParsingError, "token %s", token.String())
	}
	endToken, err := p.getNext()
	if err != nil {
		return err
	}
	switch endToken.Type() {
	case Comma:
		err = p.parsingArray(root)
		if err != nil {
			return err
		}
	case EndArray:
		return nil
	default:
		return errors.WithMessagef(ParsingError, "token: %s", endToken.String())
	}
	return nil
}
