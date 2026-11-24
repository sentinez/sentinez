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

func roundTripRules() map[string]Rule {
	return map[string]Rule{
		"all ops": {Any: [][]Cond{
			{
				{Source: SourcePath, Op: OpPrefix, Value: "/api"},
				{Source: SourceMethod, Op: OpNe, Value: "POST"},
				{Source: SourceHost, Op: OpSuffix, Value: ".com"},
			},
			{{Source: SourceIP, Op: OpEq, Value: "1.1.1.1"}},
			{{Source: SourceHeader, Key: "User-Agent", Op: OpContains,
				Value: "curl"}},
			{{Source: SourceQuery, Key: "n", Op: OpGte, Value: "10.5"}},
			{
				{Source: SourceQuery, Key: "n", Op: OpGt, Value: "1"},
				{Source: SourceQuery, Key: "n", Op: OpLt, Value: "9"},
				{Source: SourceQuery, Key: "n", Op: OpLte, Value: "8"},
			},
		}, Action: Action{Name: "block"}},
		"escapes": {Any: [][]Cond{{{
			Source: SourceHeader, Key: `X-"Q"`, Op: OpEq,
			Value: "a\\b\n\tviệt\x00",
		}}}, Action: Action{Name: "challenge"}},
		"params": {Any: [][]Cond{{
			{Source: SourcePath, Op: OpEq, Value: "/old"},
		}}, Action: Action{Name: "redirect", Params: map[string]any{
			"location": "/new", "status": float64(301), "end": true,
			"X-Tier": map[string]any{"n": float64(-1.5)},
		}}},
		"match all": {Any: [][]Cond{nil}, Action: Action{Name: "log"}},
	}
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

func TestKeyRoundTrip(t *testing.T) {
	want := [][]Cond{{{Source: SourcePath, Op: OpPrefix, Value: "/"}}}

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
			src, err := MatchSource(Rule{Any: [][]Cond{tt.give},
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
	src, err := MatchSource(Rule{Any: [][]Cond{nil},
		Action: Action{Name: "block"}})
	require.NoError(t, err)
	assert.Contains(t, src, "return SZ_BLOCK")

	src, err = MatchSource(Rule{Any: [][]Cond{nil}, Action: Action{
		Name: "redirect", Params: map[string]any{"location": "/"},
	}})
	require.NoError(t, err)
	assert.Contains(t, src, "{type = SZ_REDIRECT, params =")
}
