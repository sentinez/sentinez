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

package scripting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMatchHandWritten(t *testing.T) {
	src := `---@param r sz.Context
function match(r)
  -- admin pages, outside of the office
  if (r:path() == "/admin" or r:path():sub(1, 7) == "/admin/")
      and not (r:ip() == "10.0.0.1" or r:ip() == "10.0.0.2") then
    return {type = SZ_BLOCK}
  end
  if not (r:header("X-Debug") ~= "1") then return "block" end
  if r:ip() == "10.0.0.3" then return SZ_BLOCK end
end`

	got, err := ParseMatch(src)
	require.NoError(t, err)

	// the admin pages are one group now, not one branch per path
	want := Rule{Match: Any(
		All(
			Leaf(Cond{Source: SourceIP, Op: OpNe, Value: "10.0.0.1"}),
			Leaf(Cond{Source: SourceIP, Op: OpNe, Value: "10.0.0.2"}),
			Any(
				Leaf(Cond{Source: SourcePath, Op: OpEq, Value: "/admin"}),
				Leaf(Cond{Source: SourcePath, Op: OpPrefix,
					Value: "/admin/"}),
			),
		),
		Leaf(Cond{Source: SourceHeader, Key: "X-Debug", Op: OpEq,
			Value: "1"}),
		Leaf(Cond{Source: SourceIP, Op: OpEq, Value: "10.0.0.3"}),
	), Action: Action{Name: "block"}}

	assert.Equal(t, want, got)
}

func TestParseMatchNotDeclarative(t *testing.T) {
	tests := map[string]string{
		"other name":       `function check(ctx) return "block" end`,
		"two params":       `function match(ctx, x) return "block" end`,
		"extra stmt":       "x = 1\nfunction match(ctx) return \"block\" end",
		"no branch":        `function match(ctx) end`,
		"unknown constant": `function match(ctx) return SZ_DROP end`,
		"local":            `function match(ctx) local p = ctx:path() end`,
		"loop":             `function match(ctx) while true do end end`,
		"unknown act":      `function match(ctx) return "nuke" end`,
		"else": matchIf(`ctx:ip() == "1"`,
			`return "block" else return "log"`),
		"two actions": "function match(ctx)\n" +
			`if ctx:ip() == "1" then return "block" end` +
			"\nreturn \"log\"\nend",
	}

	for name, src := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseMatch(src)
			require.ErrorIs(t, err, ErrNotDeclarative)
		})
	}
}

func TestParseMatchReads(t *testing.T) {
	src := matchIf(`not ctx:tls() and not ctx:ip_in("10.0.0.0/8")`+
		` and not (ctx:has_query("a") or ctx:header("b") ~= "")`,
		`return SZ_BLOCK`)

	got, err := ParseMatch(src)
	require.NoError(t, err)

	want := Rule{Match: all(
		Cond{Source: SourceTLS, Op: OpNe, Value: "true"},
		Cond{Source: SourceIP, Op: OpNotWithin, Value: "10.0.0.0/8"},
		Cond{Source: SourceQuery, Key: "a", Op: OpAbsent},
		Cond{Source: SourceHeader, Key: "b", Op: OpEq},
	), Action: Action{Name: "block"}}

	assert.Equal(t, want, got)
}

func TestParseMatchSyntax(t *testing.T) {
	_, err := ParseMatch(`function match(ctx`)
	require.Error(t, err)
}

func TestParseMatchUnsupportedCondition(t *testing.T) {
	tests := map[string]string{
		"not find":    `not ctx:path():find("a", 1, true)`,
		"not prefix":  `ctx:path():sub(1, 2) ~= "/a"`,
		"sub length":  `ctx:path():sub(1, 3) == "/a"`,
		"regex find":  `ctx:path():find("a.")`,
		"var compare": `ctx:path() == ctx:host()`,
		"other ctx":   `req:path() == "/"`,
		"hex number":  `tonumber(ctx:query("n")) > 0x10`,
		"no key":      `ctx:header() == "a"`,
		"bare field":  `ctx:path()`,
		"tls compare": `ctx:tls() == "true"`,
		"tls number":  `tonumber(ctx:tls()) > 1`,
		"ip cidr":     `ctx:ip() == "10.0.0.0/8"`,
		"ip_in addr":  `ctx:ip_in("10.0.0.1")`,
		"ip_in var":   `ctx:ip_in(ctx:host())`,
		"bad regex":   `re.match(ctx:path(), "(")`,
		"regex var":   `re.match(ctx:path(), ctx:host())`,
		"regex tls":   `re.match(ctx:tls(), "t")`,
		"not regex":   `not re.match(ctx:path(), "a")`,
		"other re":    `re.find(ctx:path(), "a")`,
		"has no key":  `ctx:has_header("")`,
		"tls arg":     `ctx:tls("x")`,
	}

	for name, cond := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseMatch(matchIf(cond, `return "block"`))
			require.ErrorIs(t, err, ErrNotDeclarative)
		})
	}
}

// matchIf a match function of one if, cond then body.
func matchIf(cond, body string) string {
	return "function match(ctx)\n  if " + cond + " then\n    " + body +
		"\n  end\nend"
}

func TestParseKeyNotTrue(t *testing.T) {
	_, err := ParseKey(`function key(ctx)
		if ctx:header("X-Api-Key") ~= "" then return ctx:header("X-Api-Key") end
	end`)

	require.ErrorIs(t, err, ErrNotDeclarative)
}

func TestParseMatchTooMany(t *testing.T) {
	conds := make([]string, 0, MaxConditions+1)
	for range MaxConditions + 1 {
		conds = append(conds, `ctx:ip() == "1"`)
	}

	src := matchIf(strings.Join(conds, " and "), `return SZ_BLOCK`)

	_, err := ParseMatch(src)
	require.ErrorIs(t, err, ErrNotDeclarative)
}
