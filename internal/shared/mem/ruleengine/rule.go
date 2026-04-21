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

package ruleengine

import (
	"sync"

	corehttp "github.com/sentinez/core/http"
	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/secure/ruleengine/v1"
	"github.com/sentinez/sentinez/pkg/common/jsonx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once     sync.Once
	ruleInst *RuleCache
	mu       sync.Mutex
)

func New() *RuleCache {
	once.Do(func() {
		ruleInst = &RuleCache{
			space: ssync.NewMap[string, *ruleenginepb.RuleBased](),
		}
	})

	return ruleInst
}

func GetEngine() *RuleCache {
	return ruleInst
}

type RuleCache struct {
	space *ssync.Map[string, *ruleenginepb.RuleBased]
}

func (rc *RuleCache) Store(namespace string, gr *ruleenginepb.RuleBased) {
	val, _ := jsonx.Marshal(gr)
	zlog.Debugf("rule: load config: %s", val)

	rc.space.Store(namespace, gr)
}

func (rc *RuleCache) Load(namespace string) *ruleenginepb.RuleBased {
	expr, ok := rc.space.Load(namespace)
	if !ok {
		return nil
	}

	return expr
}

func (rc *RuleCache) LoadContext(ctx corehttp.Context) *ruleenginepb.RuleBased {
	if rc == nil {
		return nil
	}

	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge] hit rule cached %s", hCtx.GetServerName())
	return rc.Load(hCtx.GetServerName())
}

func Store(serverName string, gr *ruleenginepb.RuleBased) {
	mu.Lock()
	defer mu.Unlock()

	if ruleInst == nil {
		ruleInst = New()
	}

	ruleInst.Store(serverName, gr)
}
