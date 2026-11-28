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

package discovery

import (
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	grpcresolver "google.golang.org/grpc/resolver"
)

// fakeInstancer returns addrs, or err when set.
type fakeInstancer struct {
	mu    sync.Mutex
	addrs []string
	err   error
}

func (f *fakeInstancer) Instances(string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.addrs, f.err
}

func (f *fakeInstancer) set(addrs []string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.addrs, f.err = addrs, err
}

// fakeConn hands the states and errors of its resolver to channels.
type fakeConn struct {
	grpcresolver.ClientConn

	states chan grpcresolver.State
	errs   chan error
}

func newFakeConn() *fakeConn {
	return &fakeConn{
		states: make(chan grpcresolver.State, 1),
		errs:   make(chan error, 1),
	}
}

func (c *fakeConn) UpdateState(s grpcresolver.State) error {
	c.states <- s
	return nil
}

func (c *fakeConn) ReportError(err error) {
	c.errs <- err
}

func addrsOf(s grpcresolver.State) []string {
	var addrs []string
	for _, e := range s.Endpoints {
		for _, a := range e.Addresses {
			addrs = append(addrs, a.Addr)
		}
	}

	return addrs
}

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		require.FailNow(t, "resolver: no update")
		return *new(T)
	}
}

func build(t *testing.T, in Instancer, cc grpcresolver.ClientConn,
) grpcresolver.Resolver {
	t.Helper()

	target := grpcresolver.Target{
		URL: url.URL{Scheme: Scheme, Path: "/sentinez.apps.greeter.v1"},
	}

	// no tick in the test: lookups on Build and ResolveNow only
	r, err := NewBuilder(in, time.Hour).
		Build(target, cc, grpcresolver.BuildOptions{})
	require.NoError(t, err)

	return r
}

func TestBuilderResolves(t *testing.T) {
	in := &fakeInstancer{addrs: []string{"10.0.0.1:9000"}}
	cc := newFakeConn()
	r := build(t, in, cc)
	defer r.Close()

	assert.Equal(t, []string{"10.0.0.1:9000"}, addrsOf(receive(t, cc.states)))

	in.set([]string{"10.0.0.1:9000", "10.0.0.2:9000"}, nil)
	r.ResolveNow(grpcresolver.ResolveNowOptions{})

	assert.Equal(t, []string{"10.0.0.1:9000", "10.0.0.2:9000"},
		addrsOf(receive(t, cc.states)))
}

func TestBuilderReportsError(t *testing.T) {
	errLookup := errors.New("no service available")
	cc := newFakeConn()
	r := build(t, &fakeInstancer{err: errLookup}, cc)
	defer r.Close()

	assert.ErrorIs(t, receive(t, cc.errs), errLookup)
}

func TestTarget(t *testing.T) {
	assert.Equal(t, "consul:///sentinez.apps.iam.v1",
		Target("sentinez.apps.iam.v1"))
}
