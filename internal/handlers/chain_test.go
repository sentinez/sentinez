// Copyright 2026 Sentinéz Labs.
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

package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/sentinez/internal/dmz/edge/http/option"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/pkg/network/httpx"
	"github.com/sentinez/shared/sztracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errFake = errors.New("fake")

// fake records its calls into the shared log and passes the request on
// unless stop is set.
type fake struct {
	id           string
	log          *[]string
	stop         bool
	handleErr    error
	provisionErr error
}

func (f *fake) ID() string { return f.id }

func (f *fake) Provision(env *envplugin.Env) error {
	*f.log = append(*f.log, "provision "+f.id+" "+env.Hostname())
	return f.provisionErr
}

func (f *fake) Handle(ctx szhttp.Context, next chains.Handler) error {
	*f.log = append(*f.log, "handle "+f.id)
	if f.stop || f.handleErr != nil {
		return f.handleErr
	}

	return next.Handle(ctx)
}

func (f *fake) Cleanup() error {
	*f.log = append(*f.log, "cleanup "+f.id)
	return nil
}

func newContext() szhttp.Context {
	req := httptest.NewRequest(http.MethodGet, "https://a.example.com/", nil)
	return httpx.NewContext(req, httptest.NewRecorder())
}

func newEnv() *envplugin.Env {
	return envplugin.NewEnv(nil, option.WithHostname("example.com"))
}

func TestBuildHandle(t *testing.T) {
	tests := []struct {
		name     string
		giveStop bool // B answers the request
		wantLog  []string
	}{
		{
			name:    "runs every plugin in order",
			wantLog: []string{"handle A", "handle B", "handle C"},
		},
		{
			name:     "a plugin answering skips the rest",
			giveStop: true,
			wantLog:  []string{"handle A", "handle B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
			plugins := []chains.Plugin{
				&fake{id: "A", log: &log},
				&fake{id: "B", log: &log, stop: tt.giveStop},
				&fake{id: "C", log: &log},
			}

			chain, err := chains.Build(newEnv(), Stage, plugins...)
			require.NoError(t, err)

			log = nil
			require.NoError(t, chain.Handle(newContext()))
			assert.Equal(t, tt.wantLog, log)
		})
	}
}

func TestBuildProvision(t *testing.T) {
	var log []string

	_, err := chains.Build(newEnv(), Stage,
		&fake{id: "A", log: &log},
		&fake{id: "B", log: &log, provisionErr: errFake},
		&fake{id: "C", log: &log},
	)

	require.ErrorIs(t, err, errFake)
	assert.Equal(t, []string{
		"provision A example.com",
		"provision B example.com",
		"cleanup A",
	}, log)
}

func TestChainClose(t *testing.T) {
	var log []string
	chain, err := chains.Build(newEnv(), Stage,
		&fake{id: "A", log: &log},
		&fake{id: "B", log: &log},
	)
	require.NoError(t, err)

	log = nil
	require.NoError(t, chain.Close())
	assert.Equal(t, []string{"cleanup B", "cleanup A"}, log)
}

func TestStageError(t *testing.T) {
	var log []string
	chain, err := chains.Build(newEnv(), Stage,
		&fake{id: "A", log: &log},
		&fake{id: "B", log: &log, handleErr: errFake},
	)
	require.NoError(t, err)

	err = chain.Handle(newContext())

	require.ErrorIs(t, err, errFake)
	stage, ok := sztracer.Origin(err)
	require.True(t, ok)
	assert.Equal(t, "B", stage)
}
