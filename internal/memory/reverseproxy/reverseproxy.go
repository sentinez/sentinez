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

package reverseproxy

import (
	"sync"

	corehttp "github.com/sentinez/core/http"
	ssync "github.com/sentinez/shared/sync"
)

var (
	once             sync.Once
	reverseProxyInst *ReverseProxy
	mu               sync.Mutex
)

func New() *ReverseProxy {
	once.Do(func() {
		reverseProxyInst = &ReverseProxy{
			space: ssync.NewMap[string, corehttp.ReverseProxy](),
		}
	})

	return reverseProxyInst
}

func Get() *ReverseProxy {
	return reverseProxyInst
}

type ReverseProxy struct {
	space *ssync.Map[string, corehttp.ReverseProxy]
}

func (rp *ReverseProxy) Store(target string, rproxy corehttp.ReverseProxy) {
	rp.space.Store(target, rproxy)
}

func (rp *ReverseProxy) Load(target string) corehttp.ReverseProxy {
	expr, ok := rp.space.Load(target)
	if !ok {
		return nil
	}

	return expr
}

func Store(target string, rproxy corehttp.ReverseProxy) {
	mu.Lock()
	defer mu.Unlock()

	if reverseProxyInst == nil {
		reverseProxyInst = New()
	}

	reverseProxyInst.Store(target, rproxy)
}

func Load(target string) (corehttp.ReverseProxy, bool) {
	if reverseProxyInst == nil {
		return nil, false
	}

	return reverseProxyInst.Load(target), true
}
