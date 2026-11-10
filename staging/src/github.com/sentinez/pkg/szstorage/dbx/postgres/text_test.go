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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "valid", give: "/path?q=é", want: "/path?q=é"},
		{name: "nul", give: "a\x00b\x00", want: "a�b�"},
		{name: "invalid utf8", give: "a\xffb", want: "a�b"},
		{name: "both", give: "\x00\xc3", want: "��"},
		{name: "empty", give: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, text(tt.give))
		})
	}
}

func TestTextValue(t *testing.T) {
	tests := []struct {
		name string
		give any
		want any
	}{
		{name: "string", give: "a\x00", want: "a�"},
		{
			name: "slice",
			give: []string{"ok", "b\x00"},
			want: []string{"ok", "b�"},
		},
		{name: "valid slice", give: []string{"ok"}, want: []string{"ok"}},
		{name: "nil slice", give: []string(nil), want: []string(nil)},
		{name: "int", give: int64(7), want: int64(7)},
		{name: "bytes", give: []byte{0}, want: []byte{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, textValue(tt.give))
		})
	}
}

func TestTextValueKeepsInput(t *testing.T) {
	give := []string{"ok", "b\x00"}
	textValue(give)
	assert.Equal(t, []string{"ok", "b\x00"}, give)
}
