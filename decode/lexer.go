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
	"strconv"
)

type jsonTokenType int

func tokenToBytes(t jsonTokenType, v interface{}, ov []byte) []byte {
	if ov != nil {
		if t == STRING {
			r := []byte{'"'}
			r = append(r, ov...)
			r = append(r, '"')
			return r
		}
		return ov
	}
	switch t {
	case Boolean:
		if v.(bool) {
			return []byte{'t', 'r', 'u', 'e'}
		}
		return []byte{'f', 'a', 'l', 's', 'e'}
	case NUMBER:
		n := strconv.FormatFloat(v.(float64), 'E', -1, 64)
		return []byte(n)
	case STRING:
		var build builder
		_ = build.WriteByte('"')
		build.Write([]byte(v.(string)))
		_ = build.WriteByte('"')
		return build.Bytes()
	case NULL:
		return []byte{'n', 'u', 'l', 'l'}
	case Comma:
		return []byte{','}
	case Colon:
		return []byte{':'}
	case StartArray:
		return []byte{'['}
	case EndArray:
		return []byte{']'}
	case StartObj:
		return []byte{'{'}
	case EndObj:
		return []byte{'}'}
	case EndDoc:
		return []byte{}
	}
	return nil
}

const (
	STRING     jsonTokenType = iota + 1
	NUMBER                   // string
	NULL                     // null
	StartArray               // [
	EndArray                 // ]
	StartObj                 // {
	EndObj                   // }
	Comma                    // ,
	Colon                    // :
	Boolean                  // true false
	EndDoc
)

type jsonToken struct {
	t             jsonTokenType
	v             interface{}
	originalValue []byte
}

func (j *jsonToken) Bytes() []byte {
	return tokenToBytes(j.t, j.v, j.originalValue)
}

func (j *jsonToken) String() string {
	return string(j.Bytes())
}

type lexerTokens []*jsonToken

func (l *lexerTokens) append(t jsonTokenType, v interface{}) {
	*l = append(*l, &jsonToken{t: t, v: v})
}

func (l *lexerTokens) appendWithOV(t jsonTokenType, v interface{}, ov []byte) {
	*l = append(*l, &jsonToken{t: t, v: v, originalValue: ov})
}

type jsonParser struct {
	data         []byte
	off          int
	parserOffset int
	tokens       lexerTokens
	jsonNode     *JsonNode
}

func initLexer(data []byte) *jsonParser {
	return &jsonParser{
		data:   data,
		off:    0,
		tokens: make([]*jsonToken, 0),
	}
}

// 词法分析
func (l *jsonParser) tokenizer() error {
	data := l.data[l.off:]
	for i := 0; i < len(data); {
		b := data[i]
		switch b {
		case '{':
			l.tokens.append(StartObj, nil)
			i++
			l.off++
		case '}':
			l.tokens.append(EndObj, nil)
			i++
			l.off++
		case '[':
			l.tokens.append(StartArray, nil)
			i++
			l.off++
		case ']':
			l.tokens.append(EndArray, nil)
			i++
			l.off++
		case ':':
			l.tokens.append(Colon, nil)
			i++
			l.off++
		case ',':
			l.tokens.append(Comma, nil)
			i++
			l.off++
		case 't', 'f', 'n': // true
			err := l.tokenizerLiteral(b)
			if err != nil {
				return errors.WithStack(err)
			}
			i = (*l).off
		case '"': // string
			err := l.tokenizerString()
			if err != nil {
				return errors.WithStack(err)
			}
			i = (*l).off
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-': // number
			err := l.tokenizerNumber()
			if err != nil {
				return errors.WithStack(err)
			}
			i = (*l).off
		case '\n', ' ', '\r':
			i++
			l.off++
		default:
			return errors.WithStack(illegalInput)
		}
	}
	l.tokens.append(EndDoc, nil)
	return nil
}

// string = "" | " chars "
// chars = char | char chars
// char = any-Unicode-character-except-"-or-\-or- control-character | \" | \\ | \/ | \b | \f | \n | \r | \t | \u
func (l *jsonParser) tokenizerString() error {
	var build builder
	// stringBuff := bytes.NewBufferString("")
	l.off++
	data := l.data[l.off:]
	for i := 0; i < len(data); {
		d := data[i]
		switch d {
		case '"': // end string
			// stringBuff.WriteByte(d)
			// l.tokens.append(STRING, stringBuff.String())
			l.tokens.appendWithOV(STRING, build.String(), build.Bytes())
			l.off++
			i++
			return nil
		case '\n', '\r': // illegal input
			return errors.New("[json-diff] illegal input")
		case '\\':
			_ = build.WriteByte(d)
			if i+1 < len(data) {
				escape := data[i+1]
				// \", \\, \/, \b, \f, \n, \t, \r
				if escape == '"' || escape == '\\' || escape == '/' || escape == 'b' ||
					escape == 'f' || escape == 'n' || escape == 't' || escape == 'r' {
					_ = build.WriteByte(escape)
					i += 2
					l.off += 2
				} else {
					return illegalInput
				}
			} else {
				return illegalInput
			}

		default:
			l.off++
			i++
			_ = build.WriteByte(d)
		}
	}
	return errors.New("[json-diff] illegal input")
}

var illegalInput = errors.New("[json-diff] illegal input")

// number = int | int frac | int exp | int frac exp
// int = digit | digit1-9 digits  | - digit | - digit1-9 digits
// frac = . digits
// exp = e digits
// digits = digit | digit digits
// e = e | e+ | e-  | E | E+ | E-
func (l *jsonParser) tokenizerNumber() error {
	var build builder
	_ = build.WriteByte(l.data[l.off])
	frac := false
	exp := false
	l.off++
	data := l.data[l.off:]
	for i := 0; i < len(data); {
		d := data[i]
		switch d {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			_ = build.WriteByte(d)
			l.off++
			i++
		case '.':
			if frac || exp {
				return errors.WithStack(illegalInput)
			}
			_ = build.WriteByte(d)
			frac = true
			l.off++
			i++
		case 'e', 'E':
			if exp {
				return errors.WithStack(illegalInput)
			}
			_ = build.WriteByte(d)
			exp = true
			l.off++
			i++
			// e must be followed by a number
			if i >= len(data) {
				return errors.WithStack(illegalInput)
			}
			n := data[i]
			if n == '-' || n == '+' {
				_ = build.WriteByte(n)
				i++
				l.off++
				if i+1 >= len(l.data) {
					return errors.WithStack(illegalInput)
				}
				nn := data[i]
				if isDigits(nn) {
					_ = build.WriteByte(nn)
					l.off++
					i++
				} else {
					return errors.WithStack(illegalInput)
				}
			} else if isDigits(n) {
				_ = build.WriteByte(n)
				i++
				l.off++
			} else {
				return errors.WithStack(illegalInput)
			}
		default:
			vS := build.String()
			v, err := strconv.ParseFloat(vS, 64)
			if err != nil {
				return errors.WithStack(illegalInput)
			}
			// l.tokens.append(NUMBER, v)
			l.tokens.appendWithOV(NUMBER, v, build.Bytes())
			return nil
		}
	}
	vS := build.String()
	v, err := strconv.ParseFloat(vS, 64)
	if err != nil {
		return errors.WithStack(illegalInput)
	}
	// l.tokens.append(NUMBER, v)
	l.tokens.appendWithOV(NUMBER, v, build.Bytes())
	return nil
}

func isDigits(d byte) bool {
	return d == '0' || d == '1' || d == '2' || d == '3' ||
		d == '4' || d == '5' || d == '6' || d == '7' || d == '8' || d == '9'
}

func (l *jsonParser) literalJudge(bf *builder, lit []byte) error {
	literalSize := len(lit)
	if l.off+literalSize > len(l.data) {
		return illegalInput
	}
	get := l.data[l.off : l.off+literalSize]
	if len(lit) != len(get) {
		return errors.WithStack(illegalInput)
	}
	for i, b := range lit {
		if get[i] != b {
			return errors.WithStack(illegalInput)
		}
	}
	l.off += literalSize
	bf.Write(get)
	return nil
}

func (l *jsonParser) tokenizerLiteral(head byte) error {
	literalBuffer := &builder{}
	switch head {
	case 'n':
		err := l.literalJudge(literalBuffer, []byte{'n', 'u', 'l', 'l'})
		if err != nil {
			return err
		}
		l.tokens.append(NULL, nil)
	case 'f':
		err := l.literalJudge(literalBuffer, []byte{'f', 'a', 'l', 's', 'e'})
		if err != nil {
			return err
		}
		l.tokens.append(Boolean, false)
	case 't':
		err := l.literalJudge(literalBuffer, []byte{'t', 'r', 'u', 'e'})
		if err != nil {
			return err
		}
		l.tokens.append(Boolean, true)
	}
	return nil
}
