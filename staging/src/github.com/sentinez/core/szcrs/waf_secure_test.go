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

package szcrs

import (
	"testing"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The interruption a DetectionOnly engine would have performed is read
// through a method Coraza keeps out of types.Transaction: a Coraza upgrade
// renaming it must fail here, not silently drop the LOG events.
func TestDetectionOnlyInterruption(t *testing.T) {
	tx := denyingTransaction(t, "DetectionOnly")

	it := detectionOnlyInterruption(tx)

	require.NotNil(t, it)
	assert.Equal(t, 1, it.RuleID)
	assert.False(t, tx.IsInterrupted())
}

func TestDetectionOnlyInterruptionEngineOn(t *testing.T) {
	tx := denyingTransaction(t, "On")

	assert.Nil(t, detectionOnlyInterruption(tx))
	assert.True(t, tx.IsInterrupted())
}

// denyingTransaction runs the request headers of GET /x through a WAF of
// engine mode whose rule 1 denies it.
func denyingTransaction(t *testing.T, engine string) types.Transaction {
	t.Helper()

	waf, err := coraza.NewWAF(coraza.NewWAFConfig().WithDirectives(
		"SecRuleEngine " + engine + "\n" +
			`SecRule REQUEST_URI "@contains /x" "id:1,phase:1,deny"`,
	))
	require.NoError(t, err)

	tx := waf.NewTransaction()
	t.Cleanup(func() { assert.NoError(t, tx.Close()) })

	tx.ProcessURI("/x", "GET", "HTTP/1.1")
	tx.ProcessRequestHeaders()

	return tx
}
