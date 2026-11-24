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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// all and any the groups of the conditions cs.
func all(cs ...Cond) Node {
	nodes := make([]Node, 0, len(cs))
	for _, c := range cs {
		nodes = append(nodes, Leaf(c))
	}

	return All(nodes...)
}

// anyOf the or group of the branches, each of them the and group of its
// conditions.
func anyOf(branches ...[]Cond) Node {
	nodes := make([]Node, 0, len(branches))
	for _, b := range branches {
		nodes = append(nodes, all(b...))
	}

	return Any(nodes...)
}

func roundTripRules() map[string]Rule {
	return map[string]Rule{
		"all ops": {Match: anyOf(
			[]Cond{
				{Source: SourcePath, Op: OpPrefix, Value: "/api"},
				{Source: SourceMethod, Op: OpNe, Value: "POST"},
				{Source: SourceHost, Op: OpSuffix, Value: ".com"},
			},
			[]Cond{{Source: SourceIP, Op: OpEq, Value: "1.1.1.1"}},
			[]Cond{{Source: SourceHeader, Key: "User-Agent",
				Op: OpContains, Value: "curl"}},
			[]Cond{{Source: SourceQuery, Key: "n", Op: OpGte,
				Value: "10.5"}},
			[]Cond{
				{Source: SourceQuery, Key: "n", Op: OpGt, Value: "1"},
				{Source: SourceQuery, Key: "n", Op: OpLt, Value: "9"},
				{Source: SourceQuery, Key: "n", Op: OpLte, Value: "8"},
			},
		), Action: Action{Name: "block"}},
		"escapes": {Match: all(Cond{
			Source: SourceHeader, Key: `X-"Q"`, Op: OpEq,
			Value: "a\\b\n\tviệt\x00",
		}), Action: Action{Name: "challenge"}},
		"params": {Match: all(
			Cond{Source: SourcePath, Op: OpEq, Value: "/old"},
		), Action: Action{Name: "redirect", Params: map[string]any{
			"location": "/new", "status": float64(301), "end": true,
			"X-Tier": map[string]any{"n": float64(-1.5)},
		}}},
		"match all": {Match: All(), Action: Action{Name: "log"}},
		"reads":     readsRule(),
		"nested":    nestedRule(),
	}
}

// nestedRule (a or b) and (c or d) and not-in, a rule no list of branches
// holds without repeating its terms.
func nestedRule() Rule {
	p := func(v string) Node {
		return Leaf(Cond{Source: SourcePath, Op: OpPrefix, Value: v})
	}
	h := func(k string) Node {
		return Leaf(Cond{Source: SourceHeader, Key: k, Op: OpEq, Value: "1"})
	}
	ip := Leaf(Cond{Source: SourceIP, Op: OpNe, Value: "10.0.0.1"})

	return Rule{Match: Any(
		All(ip, Any(p("/a"), p("/b")), Any(h("X"), h("Y"), h("Z"))),
		All(h("W"), Any(All(p("/c"), Any(h("X"), p("/d"))), ip)),
	), Action: Action{Name: "block"}}
}

// readsRule a rule of the regex, CIDR, body, JA4, TLS and presence
// conditions.
func readsRule() Rule {
	return Rule{Match: anyOf(
		[]Cond{
			{Source: SourceBody, Op: OpContains, Value: "<script"},
			{Source: SourceJA4, Op: OpPrefix, Value: "t13"},
		},
		[]Cond{{Source: SourcePath, Op: OpMatches,
			Value: `^/wp-(admin|login)`}},
		[]Cond{{Source: SourceTLS, Op: OpEq, Value: "true"}},
		[]Cond{{Source: SourceTLS, Op: OpNe, Value: "true"}},
		[]Cond{{Source: SourceIP, Op: OpWithin, Value: "10.0.0.0/8"}},
		[]Cond{{Source: SourceIP, Op: OpNotWithin, Value: "::1/128"}},
		[]Cond{{Source: SourceHeader, Key: "X-Debug", Op: OpExists}},
		[]Cond{{Source: SourceQuery, Key: "debug", Op: OpAbsent}},
	), Action: Action{Name: "block"}}
}

func TestRuleRoundTrip(t *testing.T) {
	for name, want := range roundTripRules() {
		t.Run(name, func(t *testing.T) {
			src, err := MatchSource(want)
			require.NoError(t, err)

			got, err := ParseMatch(src)
			require.NoError(t, err, src)
			assert.Equal(t, want, got, src)

			p, err := Compile(name, src)
			require.NoError(t, err, src)
			assert.True(t, p.HasFunc(MatchFunc))
		})
	}
}

// TestNestedParse a hand written condition reads to the same node as the
// one written from it: and groups merged, conditions first.
func TestNestedParse(t *testing.T) {
	src := matchIf(`(ctx:path() == "/a" or ctx:path() == "/b") and `+
		`ctx:ip() == "1" and not (ctx:host() == "x" and `+
		`(ctx:method() == "GET" or ctx:method() == "PUT"))`,
		`return SZ_BLOCK`)

	got, err := ParseMatch(src)
	require.NoError(t, err)

	path := func(v string) Node {
		return Leaf(Cond{Source: SourcePath, Op: OpEq, Value: v})
	}
	method := func(v string) Node {
		return Leaf(Cond{Source: SourceMethod, Op: OpNe, Value: v})
	}
	want := All(
		Leaf(Cond{Source: SourceIP, Op: OpEq, Value: "1"}),
		Any(path("/a"), path("/b")),
		Any(Leaf(Cond{Source: SourceHost, Op: OpNe, Value: "x"}),
			All(method("GET"), method("PUT"))),
	)

	assert.Equal(t, want, got.Match)

	again, err := MatchSource(got)
	require.NoError(t, err)
	assert.Contains(t, again, `ctx:ip() == "1" and (ctx:path() == "/a" or`)
}

func TestKeyRoundTrip(t *testing.T) {
	want := all(Cond{Source: SourcePath, Op: OpPrefix, Value: "/"})

	src, err := KeySource(want)
	require.NoError(t, err)

	got, err := ParseKey(src)
	require.NoError(t, err, src)
	assert.Equal(t, want, got)
}

type sourceRunTest struct {
	name string
	give []Cond
	want bool
}

func sourceRunTests() []sourceRunTest {
	return []sourceRunTest{
		{"eq", []Cond{{Source: SourcePath, Op: OpEq, Value: "/api"}}, true},
		{"ne", []Cond{{Source: SourceMethod, Op: OpNe, Value: "GET"}},
			false},
		{"prefix", []Cond{{Source: SourcePath, Op: OpPrefix, Value: "/a"}},
			true},
		{"suffix", []Cond{{Source: SourceHost, Op: OpSuffix,
			Value: ".org"}}, false},
		{"contains", []Cond{{Source: SourceHost, Op: OpContains,
			Value: "ample"}}, true},
		{"gt", []Cond{{Source: SourceHeader, Key: "X-Debug", Op: OpGt,
			Value: "0.5"}}, true},
		{"body", []Cond{{Source: SourceBody, Op: OpContains,
			Value: `"a"`}}, true},
		{"ja4", []Cond{{Source: SourceJA4, Op: OpEq,
			Value: "t13d1516h2"}}, true},
		{"tls", []Cond{{Source: SourceTLS, Op: OpEq, Value: "false"}},
			false},
		{"matches", []Cond{{Source: SourcePath, Op: OpMatches,
			Value: "^/a.i$"}}, true},
		{"within", []Cond{{Source: SourceIP, Op: OpWithin,
			Value: "10.0.0.0/8"}}, true},
		{"not within", []Cond{{Source: SourceIP, Op: OpNotWithin,
			Value: "10.0.0.0/8"}}, false},
		{"exists", []Cond{{Source: SourceHeader, Key: "X-Debug",
			Op: OpExists}}, true},
		{"absent", []Cond{{Source: SourceQuery, Key: "q", Op: OpAbsent}},
			false},
		{"and", []Cond{
			{Source: SourceIP, Op: OpEq, Value: "10.0.0.1"},
			{Source: SourceQuery, Key: "k", Op: OpEq, Value: "q-x"},
		}, false},
	}
}

// TestRuleSourceRuns the generated sources evaluate as their rule does,
// on the fake request: GET example.com/api from 10.0.0.1, X-Debug: 1.
func TestRuleSourceRuns(t *testing.T) {
	tests := sourceRunTests()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, err := MatchSource(Rule{Match: all(tt.give...),
				Action: Action{Name: "block"}})
			require.NoError(t, err)

			p, err := Compile(tt.name, src)
			require.NoError(t, err)

			v, err := p.Invoke(context.Background(), MatchFunc, newFake())
			require.NoError(t, err)
			assert.Equal(t, tt.want, v.Kind == KindString, src)
		})
	}
}

func TestMatchSourceReturnsConstant(t *testing.T) {
	src, err := MatchSource(Rule{Match: All(),
		Action: Action{Name: "block"}})
	require.NoError(t, err)
	assert.Contains(t, src, "return SZ_BLOCK")

	src, err = MatchSource(Rule{Match: All(), Action: Action{
		Name: "redirect", Params: map[string]any{"location": "/"},
	}})
	require.NoError(t, err)
	assert.Contains(t, src, "{type = SZ_REDIRECT, params =")
}

func TestCondSource(t *testing.T) {
	path := Cond{Source: SourcePath, Op: OpEq, Value: "/a"}
	debug := Cond{Source: SourceHeader, Key: "X-Debug", Op: OpNe, Value: ""}

	tests := []struct {
		name    string
		give    Node
		want    string
		wantErr error
	}{
		{"and or", anyOf([]Cond{path, debug}, []Cond{path}),
			`ctx:path() == "/a" and ctx:header("X-Debug") ~= ""` +
				` or ctx:path() == "/a"`, nil},
		{"match all", anyOf([]Cond{path}, nil), "true", nil},
		{"nested", All(Leaf(path), Any(Leaf(debug), Leaf(path))),
			`ctx:path() == "/a" and (ctx:header("X-Debug") ~= ""` +
				` or ctx:path() == "/a")`, nil},
		{"no branch", Any(), "", errNoBranch},
		{"nested no branch", All(Leaf(path), Any()), "", errNoBranch},
		{"not declarative", all(Cond{Source: SourceHeader, Op: OpEq}),
			"", ErrNotDeclarative},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CondSource(tt.give)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCondSourceReads(t *testing.T) {
	got, err := CondSource(all(
		Cond{Source: SourcePath, Op: OpMatches, Value: "^/a"},
		Cond{Source: SourceTLS, Op: OpEq, Value: "false"},
		Cond{Source: SourceIP, Op: OpWithin, Value: "10.0.0.0/8"},
		Cond{Source: SourceHeader, Key: "X-A", Op: OpAbsent},
	))
	require.NoError(t, err)
	assert.Equal(t, `re.match(ctx:path(), "^/a") and not ctx:tls()`+
		` and ctx:ip_in("10.0.0.0/8") and not ctx:has_header("X-A")`, got)
}

func TestCondSourceInvalid(t *testing.T) {
	tests := map[string]Cond{
		"bad regex":     {Source: SourcePath, Op: OpMatches, Value: "("},
		"empty regex":   {Source: SourcePath, Op: OpMatches},
		"tls value":     {Source: SourceTLS, Op: OpEq, Value: "yes"},
		"tls op":        {Source: SourceTLS, Op: OpContains, Value: "true"},
		"bad cidr":      {Source: SourceIP, Op: OpWithin, Value: "10.0.0.1"},
		"within path":   {Source: SourcePath, Op: OpWithin, Value: "::/0"},
		"ip eq cidr":    {Source: SourceIP, Op: OpEq, Value: "10.0.0.0/8"},
		"ip prefix":     {Source: SourceIP, Op: OpPrefix, Value: "10."},
		"exists path":   {Source: SourcePath, Op: OpExists},
		"exists no key": {Source: SourceHeader, Op: OpExists},
		"exists value": {Source: SourceHeader, Key: "A", Op: OpExists,
			Value: "1"},
	}

	for name, c := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := CondSource(all(c))
			require.ErrorIs(t, err, ErrNotDeclarative)
		})
	}
}
