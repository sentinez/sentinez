// Copyright 2025 Sentinéz Labs.
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

// Package upstream keeps the reverse proxies of the upstreams, by
// namespace: each namespace dials its upstreams with its own TLS settings
// (CAs and client certificate), so a proxy is never shared between two.
package upstream

import (
	"sync"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szsync"
)

var (
	_once             sync.Once
	_reverseProxyInst *ReverseProxy
)

// New returns the process-wide reverse proxy store.
func New() *ReverseProxy {
	_once.Do(func() {
		_reverseProxyInst = &ReverseProxy{
			space: szsync.NewMap[string, map[string]szhttp.ReverseProxy](),
		}
	})

	return _reverseProxyInst
}

// ReverseProxy is a concurrent map of namespace to the proxies of its
// upstreams, by upstream server. The map of a namespace is never changed
// once stored, only replaced.
type ReverseProxy struct {
	space *szsync.Map[string, map[string]szhttp.ReverseProxy]
}

// Store sets the proxies of namespace ns, by upstream server, in place of
// the previous ones.
func (rp *ReverseProxy) Store(ns string,
	proxies map[string]szhttp.ReverseProxy) {
	rp.space.Store(ns, proxies)
}

// Load returns the proxy of namespace ns to its upstream server target.
func (rp *ReverseProxy) Load(ns, target string) (szhttp.ReverseProxy, bool) {
	proxies, ok := rp.space.Load(ns)
	if !ok {
		return nil, false
	}

	proxy, ok := proxies[target]

	return proxy, ok
}

// Delete drops the proxies of namespace ns.
func (rp *ReverseProxy) Delete(ns string) {
	rp.space.Delete(ns)
}
