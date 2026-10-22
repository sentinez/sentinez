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

package szcrs_test

import (
	"context"
	"testing"

	"github.com/corazawaf/coraza/v3"
	"github.com/sentinez/core/szcrs"
	"github.com/sentinez/core/szhttp/httpreq"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const _evalDirectives = `
SecRuleEngine On
SecRequestBodyAccess On
SecRule ARGS:q "@contains attack" "id:100,phase:1,deny,status:403"
SecRule REQUEST_HEADERS:X-Probe "@streq 1" "id:101,phase:1,pass,log"
`

// nolint:funlen
func TestEvaluate(t *testing.T) {
	waf, err := coraza.NewWAF(
		coraza.NewWAFConfig().WithDirectives(_evalDirectives))
	require.NoError(t, err)

	tests := []struct {
		name            string
		giveURI         string
		giveHeaders     []*httppb.RequestHeader
		wantRuleIds     []int
		wantInterrupted int
	}{
		{
			name:    "clean request",
			giveURI: "/?q=hello",
		},
		{
			name:    "pass rule matches without interruption",
			giveURI: "/?q=hello",
			giveHeaders: []*httppb.RequestHeader{{
				Key: []byte("X-Probe"), Values: [][]byte{[]byte("1")},
			}},
			wantRuleIds: []int{101},
		},
		{
			name:            "deny rule interrupts",
			giveURI:         "/?q=attack",
			wantRuleIds:     []int{100},
			wantInterrupted: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httpreq.NewRequestContext(
				context.Background(), &httppb.Request{
					Method:        "GET",
					Uri:           []byte(tt.giveURI),
					Protocol:      "HTTP/1.1",
					Host:          "example.com",
					RemoteAddress: "10.0.0.1:4321",
					Headers:       tt.giveHeaders,
				})

			eval, err := szcrs.Evaluate(waf, req)
			require.NoError(t, err)

			var ids []int
			for _, mr := range eval.Matched {
				ids = append(ids, mr.Rule().ID())
			}

			var interrupted int
			if eval.Interruption != nil {
				interrupted = eval.Interruption.RuleID
			}

			assert.Equal(t, tt.wantRuleIds, ids)
			assert.Equal(t, tt.wantInterrupted, interrupted)
		})
	}
}
