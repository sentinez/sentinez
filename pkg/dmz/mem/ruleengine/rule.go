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

	ruleenginepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/rule/engine/v1"
	corehttp "github.com/sentinez/sentinez/core/http"
	ssync "github.com/sentinez/sentinez/shared/sync"
	"github.com/sentinez/sentinez/shared/zlog"
)

var (
	once     sync.Once
	ruleInst *RuleCache
)

func New() *RuleCache {
	once.Do(func() {
		ruleInst = &RuleCache{
			space: ssync.NewMap[string, *ruleenginepb.Expr](),
		}
	})

	return ruleInst
}

func GetEngine() *RuleCache {
	return ruleInst
}

type RuleCache struct {
	space *ssync.Map[string, *ruleenginepb.Expr]
}

func (rc *RuleCache) Store(namespace string, expr *ruleenginepb.Expr) {
	rc.space.Store(namespace, expr)
}

func (rc *RuleCache) Load(namespace string) *ruleenginepb.Expr {
	expr, ok := rc.space.Load(namespace)
	if !ok {
		return nil
	}

	return expr
}

func (rc *RuleCache) LoadContext(ctx corehttp.Context) *ruleenginepb.Expr {
	if rc == nil {
		return nil
	}

	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge] hit rule cached %s", hCtx.GetTenantNs())
	return rc.Load(hCtx.GetTenantNs())
}
