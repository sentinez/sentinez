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

package runner

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"
)

type fakeServer struct {
	name string
	done chan struct{}
}

// recorder logs lifecycle events of fake servers, safe for concurrent use.
type recorder struct {
	mu     sync.Mutex
	events []string
}

func (r *recorder) add(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = append(r.events, event)
}

func (r *recorder) get() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Clone(r.events)
}

// fakeApp serves until stopped, or returns serveErr right away when set.
func fakeApp(name string, rec *recorder, serveErr error) *App {
	return NewApp(func(c *Context[fakeServer]) {
		c.Inject(func() *fakeServer {
			return &fakeServer{name: name, done: make(chan struct{})}
		})

		c.OnStart(func(_ context.Context, s *fakeServer) error {
			rec.add("start " + s.name)
			return nil
		})

		c.Serve(func(_ context.Context, s *fakeServer) error {
			rec.add("serve " + s.name)
			if serveErr != nil {
				return serveErr
			}

			<-s.done
			return http.ErrServerClosed
		})

		c.OnStop(func(_ context.Context, s *fakeServer) error {
			rec.add("stop " + s.name)
			close(s.done)
			return nil
		})
	})
}

func TestRunStopsAllAppsOnCancel(t *testing.T) {
	var rec recorder
	ctx, cancel := context.WithCancel(context.Background())

	errc := make(chan error, 1)
	go func() {
		errc <- (&Runner{}).Run(ctx,
			fakeApp("a", &rec, nil), fakeApp("b", &rec, nil))
	}()

	waitEvents(t, &rec, 4)
	cancel()

	if err := <-errc; err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := rec.get()
	wantFirst := []string{"start a", "start b"}
	if !slices.Equal(got[:2], wantFirst) {
		t.Errorf("start order = %v, want %v", got[:2], wantFirst)
	}

	wantLast := []string{"stop b", "stop a"}
	if !slices.Equal(got[4:], wantLast) {
		t.Errorf("stop order = %v, want %v", got[4:], wantLast)
	}
}

func TestRunStopsAllAppsWhenServeFails(t *testing.T) {
	var rec recorder
	errBoom := errors.New("boom")

	err := (&Runner{}).Run(context.Background(),
		fakeApp("a", &rec, nil), fakeApp("b", &rec, errBoom))
	if !errors.Is(err, errBoom) {
		t.Fatalf("Run err = %v, want %v", err, errBoom)
	}

	got := rec.get()
	for _, want := range []string{"stop a", "stop b"} {
		if !slices.Contains(got, want) {
			t.Errorf("events %v missing %q", got, want)
		}
	}
}

func TestRunStopsStartedAppsWhenStartFails(t *testing.T) {
	var rec recorder
	broken := NewApp(func(c *Context[fakeServer]) {
		c.Serve(func(context.Context, *fakeServer) error { return nil })
	})

	err := (&Runner{}).Run(context.Background(),
		fakeApp("a", &rec, nil), broken)
	if err == nil {
		t.Fatal("Run err = nil, want missing provider error")
	}

	want := []string{"start a", "stop a"}
	if got := rec.get(); !slices.Equal(got, want) {
		t.Errorf("events = %v, want %v", got, want)
	}
}

func waitEvents(t *testing.T, rec *recorder, n int) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for len(rec.get()) < n {
		if time.Now().After(deadline) {
			t.Fatalf("events = %v, want %d events", rec.get(), n)
		}

		time.Sleep(time.Millisecond)
	}
}
