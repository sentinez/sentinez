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

// Package httpegress pools the http egress events of the collector.
package httpegress

import (
	"github.com/sentinez/controlplane/pkg/pools"
	httppb "github.com/sentinez/sentinez/api/proto/sentinez/types/net/http/v1"
)

var _pool = pools.New(httppb.EgressEventFromVTPool)

// Pool returns the pool of the httppb.EgressEvent.
func Pool() *pools.Pool[httppb.EgressEvent, *httppb.EgressEvent] {
	return _pool
}

// Acquire returns a cleared httppb.EgressEvent.
func Acquire() *httppb.EgressEvent {
	return _pool.Acquire()
}

// Release resets obj and puts it back in the pool.
func Release(obj *httppb.EgressEvent) {
	_pool.Release(obj)
}
