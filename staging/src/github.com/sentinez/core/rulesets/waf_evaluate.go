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

package corers

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/shared/bytesconv"
)

// Evaluation is the outcome of the request phases of a WAF.
type Evaluation struct {
	// Interruption is the disruptive action taken, nil if the request
	// passes.
	Interruption *types.Interruption
	Matched      []types.MatchedRule
}

// Evaluate runs the request phases (connection, headers, body) of waf
// against req. Unlike Rulesets it needs no live connection, so a
// recorded request can be replayed against any setting.
func Evaluate(waf coraza.WAF, req corehttp.Request) (Evaluation, error) {
	tx := newTransaction(waf, req)
	defer func() { _ = tx.Close() }()

	it, err := evaluateRequest(tx, req)
	if err != nil {
		return Evaluation{}, err
	}

	tx.ProcessLogging()

	// tx goes back to the coraza pool on Close: copy what it holds.
	return Evaluation{
		Interruption: it,
		Matched:      slices.Clone(tx.MatchedRules()),
	}, nil
}

func evaluateRequest(
	tx types.Transaction, req corehttp.Request,
) (*types.Interruption, error) {
	client, port := splitAddr(req.RemoteAddr())
	tx.ProcessConnection(client, port, "", 0)
	tx.ProcessURI(
		bytesconv.B2s(req.URI()), bytesconv.B2s(req.Method()), req.Protocol())

	for k, values := range req.Headers() {
		for _, v := range values {
			tx.AddRequestHeader(k, bytesconv.B2s(v))
		}
	}

	if host := bytesconv.B2s(req.Host()); host != "" {
		tx.AddRequestHeader("Host", host)
		tx.SetServerName(host)
	}

	if it := tx.ProcessRequestHeaders(); it != nil {
		return it, nil
	}

	if body := req.Body(); tx.IsRequestBodyAccessible() && len(body) != 0 {
		it, _, err := tx.ReadRequestBodyFrom(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}

		if it != nil {
			return it, nil
		}
	}

	return tx.ProcessRequestBody()
}
