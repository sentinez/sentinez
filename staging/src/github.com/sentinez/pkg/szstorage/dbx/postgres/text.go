// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres

import (
	"strings"
	"unicode/utf8"
)

// _replacement stands for the bytes a postgres text cannot hold.
const _replacement = "�"

// textValue returns v, a string or a []string, as text postgres accepts:
// the NUL bytes and invalid UTF-8 sequences, which fail the statement
// (SQLSTATE 22021), become U+FFFD. The other values are returned as is.
func textValue(v any) any {
	switch t := v.(type) {
	case string:
		return text(t)
	case []string:
		return textSlice(t)
	default:
		return v
	}
}

// text returns s with its NUL bytes and invalid UTF-8 sequences replaced
// by U+FFFD; s itself when it has none.
func text(s string) string {
	if utf8.ValidString(s) && strings.IndexByte(s, 0) < 0 {
		return s
	}

	return strings.ReplaceAll(
		strings.ToValidUTF8(s, _replacement), "\x00", _replacement)
}

// textSlice returns s with each element made text; s itself, not
// copied, when none changes.
func textSlice(s []string) []string {
	var out []string
	for i, v := range s {
		t := text(v)
		if out == nil && t != v {
			out = make([]string, len(s))
			copy(out, s[:i])
		}
		if out != nil {
			out[i] = t
		}
	}
	if out == nil {
		return s
	}

	return out
}
