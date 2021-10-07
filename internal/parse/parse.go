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

type JsonParser struct {
	raw     []byte
	lexer   *Lexer
	parsing *Parsing
	// tokenCh chan *Token
	tokenList []*Token
	tokenPool *TokenPool
	// sChan   chan struct{} // stop
	// err     error
}

func NewParser() *JsonParser {
	p := &JsonParser{}
	p.lexer = NewLexer(p)
	p.parsing = NewParsing(p)
	// p.sChan = make(chan struct{})
	// p.tokenCh = make(chan *Token)
	p.tokenList = make([]*Token, 0)
	p.tokenPool = NewTokenPool(10)
	return p
}

func (p *JsonParser) Parser(data []byte) (json_node.IJsonNode, error) {

	p.raw = data
	err := p.lexer.Scanning()
	if err != nil {
		return nil, err
	}
	// fmt.Println(len(data), "---------",  len(p.tokenList))
	err = p.parsing.Scanner()
	if err != nil {
		return nil, err
	}
	return p.parsing.node, nil
}
