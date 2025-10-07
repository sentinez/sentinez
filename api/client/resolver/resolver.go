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

package resolver

import (
	"errors"
	"sync"
	"time"

	"github.com/sentinez/sentinez/api/client/consul"
	"github.com/sony/gobreaker"
)

type ServiceDiscovery interface {
	Discover(name string) ([]consul.Instance, error)
}

type Resolver struct {
	discovery ServiceDiscovery
	cache     map[string][]consul.Instance
	lastFetch map[string]time.Time
	breakers  map[string]*gobreaker.CircuitBreaker
	mu        sync.RWMutex
	index     map[string]int
}

func New(sd ServiceDiscovery) *Resolver {
	return &Resolver{
		discovery: sd,
		cache:     make(map[string][]consul.Instance),
		lastFetch: make(map[string]time.Time),
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		index:     make(map[string]int),
	}
}

func (r *Resolver) GetInstances(name string) ([]consul.Instance, error) {
	r.mu.RLock()
	instances, ok := r.cache[name]
	last := r.lastFetch[name]
	r.mu.RUnlock()

	if ok && time.Since(last) < 10*time.Second {
		return instances, nil
	}

	newInstances, err := r.discovery.Discover(name)
	if err != nil || len(newInstances) == 0 {
		return nil, errors.New("no service available")
	}

	r.mu.Lock()
	r.cache[name] = newInstances
	r.lastFetch[name] = time.Now()
	r.mu.Unlock()

	return newInstances, nil
}

func (r *Resolver) PickInstance(name string) (*consul.Instance, error) {
	instances, err := r.GetInstances(name)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(instances) == 0 {
		return nil, errors.New("no instance available")
	}

	idx := r.index[name] % len(instances)
	r.index[name] = (r.index[name] + 1) % len(instances)
	inst := instances[idx]

	return &inst, nil
}

func (r *Resolver) GetBreaker(key string) *gobreaker.CircuitBreaker {
	r.mu.Lock()
	defer r.mu.Unlock()

	if b, ok := r.breakers[key]; ok {
		return b
	}

	b := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        key,
		MaxRequests: 1,
		Timeout:     5 * time.Second,
		Interval:    10 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures >= 3
		},
	})
	r.breakers[key] = b
	return b
}
