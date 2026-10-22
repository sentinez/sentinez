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

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/shared/szsync"
)

var (
	_once             sync.Once
	_reverseProxyInst *ReverseProxy
	_mu               sync.Mutex
)

func New() *ReverseProxy {
	_once.Do(func() {
		_reverseProxyInst = &ReverseProxy{
			space: szsync.NewMap[string, szhttp.ReverseProxy](),
		}
	})

	return _reverseProxyInst
}

type ReverseProxy struct {
	space *szsync.Map[string, szhttp.ReverseProxy]
}

func (rp *ReverseProxy) Store(target string, rproxy szhttp.ReverseProxy) {
	rp.space.Store(target, rproxy)
}

func (rp *ReverseProxy) Load(target string) (szhttp.ReverseProxy, bool) {
	expr, ok := rp.space.Load(target)
	if !ok {
		return nil, false
	}

	return expr, true
}
