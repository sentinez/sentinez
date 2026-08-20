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

package wafengine

import (
	"sync"

	"github.com/corazawaf/coraza/v3"
	corehttp "github.com/sentinez/core/http"
	corers "github.com/sentinez/core/rulesets"
	"github.com/sentinez/sentinez"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once    sync.Once
	wafInst *WAFRulesets
)

func New() *WAFRulesets {
	once.Do(func() {
		wafInst = &WAFRulesets{
			space: ssync.NewMap[string, coraza.WAF](),
		}
	})

	return wafInst
}

type WAFRulesets struct {
	// key: namespace
	// ex: dev.sentinez.vn
	//	- domain: sentinez.vn
	// 	- namespace: dev
	space *ssync.Map[string, coraza.WAF]
}

func (w *WAFRulesets) Store(namespace string, flag corers.Flag) error {

	if w == nil {
		return nil
	}

	version, fs := sentinez.WAF4160()

	waf, err := corers.NewWAF(version, fs, flag)
	if err != nil {
		return err
	}

	w.space.Store(namespace, waf)
	if waf != nil {
		zlog.Infof("edge: WAF initialized successfully, ns=%s", namespace)
	}

	return nil
}

func (w *WAFRulesets) Load(namespace string) (coraza.WAF, bool) {
	if w == nil {
		return nil, false
	}

	value, ok := w.space.Load(namespace)
	if !ok {
		return nil, false
	}

	return value, true
}

func (w *WAFRulesets) LoadContext(ctx corehttp.Context) (coraza.WAF, bool) {
	if w == nil {
		return nil, false
	}

	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		return nil, false
	}

	zlog.Debugf("[edge][namespace] hit waf cached %s", hCtx.GetServerName())
	return w.Load(hCtx.GetServerName())
}
