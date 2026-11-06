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

// Package secrule pools the secrule and rate limit events of the collector.
package secrule

import (
	"github.com/sentinez/controlplane/pkg/pools"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
)

var _pool = pools.New(secrulepb.RuleEventFromVTPool)

// Pool returns the pool of the secrulepb.RuleEvent.
func Pool() *pools.Pool[secrulepb.RuleEvent, *secrulepb.RuleEvent] {
	return _pool
}

// Acquire returns a cleared secrulepb.RuleEvent.
func Acquire() *secrulepb.RuleEvent {
	return _pool.Acquire()
}

// Release resets obj and puts it back in the pool.
func Release(obj *secrulepb.RuleEvent) {
	_pool.Release(obj)
}
