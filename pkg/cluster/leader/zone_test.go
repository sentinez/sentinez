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

	"github.com/sentinez/sentinez/api/client/consul"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishWritesOnce(t *testing.T) {
	store := &fakeStore{}
	task := newZoneElector(store).Publish("setting/web", "v")

	require.NoError(t, task(context.Background(), Term{Index: 4}))

	require.Len(t, store.puts, 1)
	assert.Equal(t, "sentinez/dmz/edge/zone/sgn/setting/web",
		store.puts[0].key)
	assert.Equal(t, uint64(4), store.puts[0].token)
	assert.Equal(t, "v", store.puts[0].value)
}

func TestPublishRetriesUntilWritten(t *testing.T) {
	down := errors.New("consul down")
	store := &fakeStore{errs: []error{down, down}}
	task := newZoneElector(store).Publish("setting/web", "v")

	ctx, cancel := context.WithTimeout(context.Background(), _waitFor)
	defer cancel()

	require.NoError(t, task(ctx, Term{Index: 4}))
	assert.Len(t, store.puts, 3)
}

func TestPublishFencedEndsTerm(t *testing.T) {
	store := &fakeStore{errs: []error{consul.ErrFenced}}
	task := newZoneElector(store).Publish("setting/web", "v")

	err := task(context.Background(), Term{Index: 4})

	require.ErrorIs(t, err, consul.ErrFenced)
}

func TestPublishStandalone(t *testing.T) {
	e := newElector(&standalone{node: "edge-01"})

	assert.Nil(t, e.Publish("setting/web", "v"))
}

func TestPublishStopsWithContext(t *testing.T) {
	store := &fakeStore{errs: []error{errors.New("consul down")}}
	e := newZoneElector(store)
	e.zone.interval = time.Hour
	task := e.Publish("setting/web", "v")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, task(ctx, Term{Index: 4}))
}
