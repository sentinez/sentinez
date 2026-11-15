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

package szprotobuf

import (
	"strings"

	"github.com/sentinez/pkg/szcolor"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Diff formats, like git diff, the change of a message: the lines only Old
// has in red after a "-", those only New has in green after a "+", the
// others unmarked. A nil side is an empty one. Nothing is formatted until
// String is called, so a Diff costs nothing as a log argument whose level
// is off.
type Diff struct {
	Old, New proto.Message
}

// String returns the colored line diff.
func (d Diff) String() string {
	a, b := textLines(d.Old), textLines(d.New)

	var sb strings.Builder
	i, j := 0, 0
	for _, m := range matches(a, b) {
		for ; i < m[0]; i++ {
			writeLine(&sb, szcolor.Red, "- ", a[i])
		}

		for ; j < m[1]; j++ {
			writeLine(&sb, szcolor.Green, "+ ", b[j])
		}

		if i < len(a) && j < len(b) {
			writeLine(&sb, 0 /* no color */, "  ", a[i])
			i, j = i+1, j+1
		}
	}

	return sb.String()
}

// writeLine writes a line after its sign, colored unless c is 0, to b, on
// a new line unless it is the first.
func writeLine(b *strings.Builder, c szcolor.Color, sign, line string) {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}

	if c == 0 {
		b.WriteString(sign)
		b.WriteString(line)

		return
	}

	b.WriteString(c.Add(sign + line))
}

// textLines the fields of a message, one per line.
func textLines(m proto.Message) []string {
	if m == nil || !m.ProtoReflect().IsValid() {
		return nil
	}

	text := prototext.MarshalOptions{Multiline: true}.Format(m)

	return strings.Split(strings.TrimSpace(text), "\n")
}

// matches the index pairs of the lines a and b have in common, along their
// longest common subsequence, ending with {len(a), len(b)}. The common head
// and tail are matched as is, the dynamic programming only runs on what
// lies between, usually a few lines.
func matches(a, b []string) [][2]int {
	head := 0
	for head < len(a) && head < len(b) && a[head] == b[head] {
		head++
	}

	tail := 0
	for tail < len(a)-head && tail < len(b)-head &&
		a[len(a)-1-tail] == b[len(b)-1-tail] {
		tail++
	}

	pairs := make([][2]int, 0, head+tail+1)
	for k := range head {
		pairs = append(pairs, [2]int{k, k})
	}

	pairs = append(pairs, lcs(a[head:len(a)-tail], b[head:len(b)-tail],
		head)...)
	for k := tail; k > 0; k-- {
		pairs = append(pairs, [2]int{len(a) - k, len(b) - k})
	}

	return append(pairs, [2]int{len(a), len(b)})
}

// lcs the index pairs, shifted by off, of the longest common subsequence
// of a and b.
func lcs(a, b []string, off int) [][2]int {
	// n[i*w+j] is the length of the lcs of a[i:] and b[j:].
	w := len(b) + 1
	n := make([]int, (len(a)+1)*w)
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				n[i*w+j] = n[(i+1)*w+j+1] + 1
			} else {
				n[i*w+j] = max(n[(i+1)*w+j], n[i*w+j+1])
			}
		}
	}

	var pairs [][2]int
	for i, j := 0, 0; i < len(a) && j < len(b); {
		switch {
		case a[i] == b[j]:
			pairs = append(pairs, [2]int{off + i, off + j})
			i, j = i+1, j+1
		case n[(i+1)*w+j] >= n[i*w+j+1]:
			i++
		default:
			j++
		}
	}

	return pairs
}
