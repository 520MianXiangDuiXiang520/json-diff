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

package error

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	PathNotFindError = errors.New("path not find")
	NodeTypeError    = errors.New("bad type")
	UnEqual          = errors.New("UnEqual")
	KeyExisted       = errors.New("key existed")
)

func KeyExistedF(path string) error {
	return errors.Wrap(KeyExisted, path)
}

func PathNotFind(path string) error {
	return errors.Wrap(PathNotFindError, path)
}

func BadNodeType(msg string) error {
	return errors.Wrap(NodeTypeError, msg)
}

func UnEqualF(want, got string) error {
	return errors.Wrap(UnEqual, fmt.Sprintf("want %s, got %s", want, got))
}
