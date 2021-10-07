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
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
)

// Lexer 词法分析器
type Lexer struct {
	parser *JsonParser
	offset int
}

func (l *Lexer) Clear() {
	l.offset = 0
}

func NewLexer(p *JsonParser) *Lexer {
	return &Lexer{parser: p}
}

func (l *Lexer) Scanning() error {
	if len(l.parser.raw) <= 0 {
		return errors.WithMessage(LexerError, "no data")
	}
	p := l.parser
	for l.offset < len(p.raw) {
		h := p.raw[l.offset]
		switch h {
		case '{':
			// p.tokenList = append(p.tokenList,   p.tokenPool.Get(StartObj)
			p.tokenList = append(p.tokenList, p.tokenPool.Get(StartObj))
			l.offset++
		case '}':
			p.tokenList = append(p.tokenList, p.tokenPool.Get(EndObj))
			l.offset++
		case '[':
			p.tokenList = append(p.tokenList, p.tokenPool.Get(StartArray))
			l.offset++
		case ']':
			p.tokenList = append(p.tokenList, p.tokenPool.Get(EndArray))
			l.offset++
		case ',':
			p.tokenList = append(p.tokenList, p.tokenPool.Get(Comma))
			l.offset++
		case ':':
			p.tokenList = append(p.tokenList, p.tokenPool.Get(Colon))
			l.offset++
		case '"':
			// string = quotation-mark *char quotation-mark
			l.offset++
			err := l.scanningString()
			if err != nil {
				return err
			}
		case 0x20, 0x09, 0x0A, 0x0D:
			// Space, Horizontal tab, Line feed or New line, Carriage return
			l.offset++
		case '-':
			l.offset++
			err := l.scanningNumber(true)
			if err != nil {
				return err
			}
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			err := l.scanningNumber(false)
			if err != nil {
				return err
			}
		case 't':
			err := l.scanningValue("true")
			if err != nil {
				return err
			}
		case 'f':
			err := l.scanningValue("false")
			if err != nil {
				return nil
			}
		case 'n':
			err := l.scanningValue("null")
			if err != nil {
				return err
			}
		default:
			return errors.WithMessagef(LexerError, "offset %d: %s", l.offset, string(h))
		}

	}
	// l.parser.tokenList = append(l.parser.tokenList,  l.parser.tokenPool.Get(EndDoc)
	l.parser.tokenList = append(l.parser.tokenList, l.parser.tokenPool.Get(EndDoc))
	return nil
}

func (l *Lexer) scanningValue(v string) error {
	if l.offset+len(v) > len(l.parser.raw) {
		return errors.WithMessagef(LexerError, "offset %d ", l.offset)
	}
	g := string(l.parser.raw[l.offset : l.offset+len(v)])
	if g != v {
		return errors.WithMessagef(LexerError, "offset %d ", l.offset)
	}
	switch v {
	case "true":
		l.parser.tokenList = append(l.parser.tokenList, l.parser.tokenPool.Get(True))
	case "false":
		l.parser.tokenList = append(l.parser.tokenList, l.parser.tokenPool.Get(False))
	case "null":
		l.parser.tokenList = append(l.parser.tokenList, l.parser.tokenPool.Get(NULL))
	}
	l.offset += len(v)
	return nil
}

func (l *Lexer) scanningNumber(minus bool) error {
	p := l.parser
	startWithZero := false
	res := bytes.NewBufferString("")
	if minus {
		res.WriteByte('-')
	}

	if l.offset >= len(p.raw) {
		return errors.WithMessagef(LexerError, "offset %d: ", l.offset)
	}
	f := p.raw[l.offset]
	if f == '0' {
		startWithZero = true
	}

	if f < '0' || f > '9' {
		return errors.WithMessagef(LexerError, "offset %d: %s", l.offset, string(f))
	}
	res.WriteByte(f)
	l.offset++

	for l.offset < len(p.raw) {
		d := p.raw[l.offset]
		switch d {
		case '.':
			res.WriteByte(d)
			l.offset++
			err := l.frac(res)
			if err != nil {
				return err
			}
			token := p.tokenPool.Get(NUMBER)
			token.d = res.Bytes()
			p.tokenList = append(p.tokenList, token)
			return nil
		case 'e', 'E':
			res.WriteByte(d)
			l.offset++
			err := l.exp(res)
			if err != nil {
				return err
			}
			token := p.tokenPool.Get(NUMBER)
			token.d = res.Bytes()
			p.tokenList = append(p.tokenList, token)
			return nil
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			if startWithZero {
				return errors.WithMessagef(LexerError, "leading zeros are not allowed")
			}
			res.WriteByte(d)
			l.offset++
		default:
			token := p.tokenPool.Get(NUMBER)
			token.d = res.Bytes()
			p.tokenList = append(p.tokenList, token)
			return nil
		}
	}
	return nil
}

func (l *Lexer) frac(res *bytes.Buffer) error {
	p := l.parser
	if l.offset >= len(p.raw) {
		return errors.WithMessagef(LexerError, "offset %d: .", l.offset)
	}
	f := p.raw[l.offset]
	if f < '0' || f > '9' {
		return errors.WithMessagef(LexerError, "offset %d: %s", l.offset, string(f))
	}
	res.WriteByte(f)
	l.offset++
	for l.offset < len(p.raw) {
		d := p.raw[l.offset]
		switch d {
		case 'e', 'E':
			res.WriteByte(d)
			l.offset++
			err := l.exp(res)
			return err
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			res.WriteByte(d)
			l.offset++
		default:
			return nil
		}
	}
	return nil
}

func (l *Lexer) exp(res *bytes.Buffer) error {
	p := l.parser
	if l.offset >= len(p.raw) {
		return errors.WithMessagef(LexerError, "offset %d: E", l.offset)
	}
	f := p.raw[l.offset]
	if !((f >= '0' && f <= '9') || f == '-' || f == '+') {
		return errors.WithMessagef(LexerError, "offset %d: %s", l.offset, string(f))
	}
	res.WriteByte(f)
	l.offset++

	if f == '-' || f == '+' {
		if l.offset >= len(p.raw) {
			return errors.WithMessagef(LexerError, "offset %d: E", l.offset)
		}
		s := p.raw[l.offset]
		if s < '0' || s > '9' {
			return errors.WithMessagef(LexerError, "offset %d: %s", l.offset, string(s))
		}
		res.WriteByte(s)
		l.offset++
	}
	for l.offset < len(p.raw) {
		d := p.raw[l.offset]
		switch d {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			res.WriteByte(d)
			l.offset++
		default:
			return nil
		}
	}
	return nil
}

func (l *Lexer) scanningString() error {
	p := l.parser
	res := bytes.NewBufferString("")
	for l.offset < len(p.raw) {
		d := p.raw[l.offset]
		switch d {
		case '"': // end of string
			l.offset++
			token := p.tokenPool.Get(STRING)
			token.d = res.Bytes()
			p.tokenList = append(p.tokenList, token)
			return nil
		case '\\':
			l.offset++
			err := l.escape(res)
			if err != nil {
				fmt.Println(err)
				return err
			}
		default:
			res.WriteByte(d)
			l.offset++
		}
	}
	return errors.WithMessagef(LexerError, "offset: %d", l.offset)
}

func (l *Lexer) escape(buf *bytes.Buffer) error {
	p := l.parser
	switch p.raw[l.offset] {
	case 0x22: // "
		buf.WriteString(`\"`)
		l.offset++
	case 0x5C:
		buf.WriteString(`\\`)
		l.offset++
	case 0x2F:
		buf.WriteString("//")
		l.offset++
	case 0x62:
		buf.WriteString(`\b`)
		l.offset++
	case 0x66:
		buf.WriteString(`\f`)
		l.offset++
	case 0x6E:
		buf.WriteString(`\n`)
		l.offset++
	case 0x72:
		buf.WriteString(`\r`)
		l.offset++
	case 0x74:
		buf.WriteString(`\t`)
		l.offset++
	case 'u':
		l.offset++
		bt, err := l.unicodeBMP()
		if err != nil {
			return err
		}
		buf.WriteString(`\u`)
		buf.Write(bt)
	default:
		return errors.WithMessagef(LexerError, "offset: %d data %s", l.offset, string(p.raw[l.offset]))
	}
	return nil
}

func (l *Lexer) unicodeBMP() ([]byte, error) {
	p := l.parser
	if l.offset+4 > len(p.raw) {
		return nil, errors.WithMessagef(LexerError, "offset: %d", l.offset)
	}
	uS := bytes.NewBufferString("")
	for i := 0; i < 4; i++ {
		d := p.raw[l.offset]
		uS.WriteByte(d)
		l.offset++
	}
	_, err := hex.DecodeString(uS.String())
	if err != nil {
		return nil, errors.WithMessage(LexerError, err.Error())
	}

	return uS.Bytes(), nil
}
