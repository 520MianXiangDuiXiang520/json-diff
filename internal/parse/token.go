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
)

type TokenType uint8

const (
	STRING TokenType = iota + 1
	NUMBER
	NULL       // null
	StartArray // [
	EndArray   // ]
	StartObj   // {
	EndObj     // }
	Comma      // ,
	Colon      // :
	True
	False
	EndDoc
)

type Token struct {
	d []byte
	t TokenType
}

func (t *Token) String() string {
	switch t.t {
	case NULL:
		return "null"
	case True:
		return "true"
	case False:
		return "false"
	case StartArray:
		return "["
	case StartObj:
		return "{"
	case EndArray:
		return "]"
	case EndObj:
		return "}"
	case EndDoc:
		return "#END"
	case Comma:
		return ","
	case Colon:
		return ":"
	}
	return string(t.d)
}

func (t *Token) Bytes() []byte {
	return t.d
}

func (t *Token) Type() TokenType {
	return t.t
}

type TokenPool struct {
	nullC   *Token
	trueC   *Token
	falseC  *Token
	leftAC  *Token // [
	rightAC *Token // ]
	leftOC  *Token // {
	rightOC *Token // }
	commaC  *Token // ,
	colonC  *Token // :
	endC    *Token
}

func NewTokenPool(size int) *TokenPool {
	return &TokenPool{
		nullC:   &Token{t: NULL},
		trueC:   &Token{t: True},
		falseC:  &Token{t: False},
		leftAC:  &Token{t: StartArray},
		rightAC: &Token{t: EndArray},
		leftOC:  &Token{t: StartObj},
		rightOC: &Token{t: EndObj},
		commaC:  &Token{t: Comma},
		colonC:  &Token{t: Colon},
		endC:    &Token{t: EndDoc},
	}
}

func (p *TokenPool) Get(t TokenType) *Token {
	switch t {
	case NULL:
		return p.nullC
	case True:
		return p.trueC
	case False:
		return p.falseC
	case StartArray:
		return p.leftAC
	case StartObj:
		return p.leftOC
	case EndArray:
		return p.rightAC
	case EndObj:
		return p.rightOC
	case EndDoc:
		return p.endC
	case Comma:
		return p.commaC
	case Colon:
		return p.colonC
	case STRING:
		return &Token{t: STRING}
	case NUMBER:
		return &Token{t: NUMBER}
	}
	return nil
}

func getNodeBySimpleToken(t *Token) json_node.IJsonNode {
	switch t.Type() {
	case NULL:
		return json_node.NewNullNode()
	case STRING:
		return json_node.NewStringNode(t.Bytes())
	case NUMBER:
		return json_node.NewNumberNode(t.Bytes())
	case False:
		return json_node.NewBooleanNode(false)
	case True:
		return json_node.NewBooleanNode(true)
	}
	return nil
}
