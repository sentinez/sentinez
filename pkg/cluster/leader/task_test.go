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

package leader

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllSkipsNil(t *testing.T) {
	assert.Nil(t, All())
	assert.Nil(t, All(nil, nil))
}

func TestAllFailureCancelsOthers(t *testing.T) {
	boom := errors.New("boom")
	waiting := func(ctx context.Context, _ Term) error {
		<-ctx.Done()
		return nil
	}
	failing := func(context.Context, Term) error { return boom }

	err := All(waiting, nil, failing)(context.Background(), Term{})

	require.ErrorIs(t, err, boom)
}

func TestAllDoneLeavesOthersRunning(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	done := func(context.Context, Term) error { return nil }
	waiting := func(ctx context.Context, _ Term) error {
		close(started)
		<-ctx.Done()
		return nil
	}

	errCh := make(chan error, 1)
	go func() { errCh <- All(done, waiting)(ctx, Term{}) }()
	<-started

	select {
	case <-errCh:
		t.Fatal("All returned while a task still runs")
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	require.NoError(t, <-errCh)
}
