// Copyright 2025 Duc-Hung Ho.
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

// Package color provides text coloring.
package color

import (
	"fmt"
)

// Foreground colors.
const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Color represents a text color.
type Color uint8

// Add adds the coloring to the given string.
func (c Color) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

// Status returns the colored status code based on the first digit of the code.
func Status(code int) string {
	reason := code / 100
	resp := fmt.Sprintf("%d", code)
	switch reason {
	case 2:
		return Green.Add(resp)
	case 3:
		return Yellow.Add(resp)
	case 4:
		return Red.Add(resp)
	case 5:
		return Red.Add(resp)
	default:
		return White.Add(resp)
	}
}
