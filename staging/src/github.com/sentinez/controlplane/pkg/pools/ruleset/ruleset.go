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

// Package ruleset pools the ruleset events of the collector.
package ruleset

import (
	"github.com/sentinez/controlplane/pkg/pools"
	rulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/ruleset/v1"
)

var _pool = pools.New(rulesetpb.RulesetEventFromVTPool)

// Pool returns the pool of the rulesetpb.RulesetEvent.
func Pool() *pools.Pool[rulesetpb.RulesetEvent, *rulesetpb.RulesetEvent] {
	return _pool
}

// Acquire returns a cleared rulesetpb.RulesetEvent.
func Acquire() *rulesetpb.RulesetEvent {
	return _pool.Acquire()
}

// Release resets obj and puts it back in the pool.
func Release(obj *rulesetpb.RulesetEvent) {
	_pool.Release(obj)
}
