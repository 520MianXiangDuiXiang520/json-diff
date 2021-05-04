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

package json_diff

import (
    `fmt`
    `github.com/pkg/errors`
)

// BadDiffsError 在输入不合法的 diffs 串时被返回
var BadDiffsError = errors.New("diffs format error")

type JsonNodeError struct {
    Op string
}

func (e *JsonNodeError) Error() string {
    return fmt.Sprintf("fail to %s", e.Op)
}

func GetJsonNodeError(op, msg string) error {
    return errors.Wrap(&JsonNodeError{Op: op}, msg)
}

func WrapJsonNodeError(op string, err error) error {
    return errors.Wrap(err, fmt.Sprintf("fail to %s", op))
}