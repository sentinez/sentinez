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

//go:build ruleguard

// Package lint holds the ruleguard rules of the project, run by gocritic
// from .golangci.yaml. The ruleguard tag keeps the file out of the builds;
// go.mod requires its dsl import.
package lint

import "github.com/quasilyte/go-ruleguard/dsl"

// _cheapCall the calls allowed in the arguments of a debug log: getters and
// builtins that do not allocate.
const _cheapCall = `(^|\.)(Get[A-Z]\w*|RequestId|Unwrap|NetConn|Path)$|^(len|cap|time\.Since)$`

// lazyDebug: the arguments of szlog.Debug/Debugf are evaluated even when the
// debug level is off, so they must not call anything costly. Pass a value
// whose String() builds the text, or use szlog.DebugFn.
func lazyDebug(m dsl.Matcher) {
	m.Import("github.com/sentinez/pkg/szlog")

	m.Match(
		`szlog.Debugf($_, $*_, $f($*_), $*_)`,
		`szlog.Debug($*_, $f($*_), $*_)`,
	).
		Where(!m["f"].Text.Matches(_cheapCall)).
		Report(`debug log argument calls $f even when debug is off: ` +
			`pass a fmt.Stringer or use szlog.DebugFn`)
}

// colorOnlyInDebug: the color escape codes end up in the files and OTLP
// that the other levels go to.
func colorOnlyInDebug(m dsl.Matcher) {
	m.Import("github.com/sentinez/pkg/szlog")
	m.Import("github.com/sentinez/pkg/szcolor")

	m.Match(
		`szlog.$lvl($*_, szcolor.$_.Add($*_), $*_)`,
		`szlog.$lvl($*_, szcolor.Status($*_), $*_)`,
	).
		Where(!m["lvl"].Text.Matches(`^Debug`)).
		Report(`szcolor in a szlog.$lvl message: colors are for debug only`)
}

// builderConcat: writing a concatenation to a builder allocates the
// concatenated string first; write the pieces.
func builderConcat(m dsl.Matcher) {
	m.Match(`$b.WriteString($x + $y)`).
		Where(m["b"].Type.Is(`strings.Builder`) ||
			m["b"].Type.Is(`*strings.Builder`) ||
			m["b"].Type.Is(`bytes.Buffer`) ||
			m["b"].Type.Is(`*bytes.Buffer`)).
		Report(`write $x and $y to $b one after the other`)
}
