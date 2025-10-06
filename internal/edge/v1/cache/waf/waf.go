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

package wafcache

import (
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/sentinez/sentinez/pkg/syncx"
	"github.com/sentinez/sentinez/rules"
)

var (
	once    sync.Once
	wafInst *WAFCache
)

func New() *WAFCache {
	once.Do(func() {
		wafInst = &WAFCache{
			wafMap: syncx.NewMap[string, coraza.WAF](),
		}
	})

	return wafInst
}

func GetWafCache() *WAFCache {
	return wafInst
}

type WAFCache struct {
	// key: namespace
	// ex: dev.sentinez.vn
	//	- domain: sentinez.vn
	// 	- namespace: dev
	wafMap *syncx.Map[string, coraza.WAF]
}

func (w *WAFCache) Store(conf *common.AppConfig, namespace string,
	version rules.Version, flag rules.RulesetsFlag) error {

	waf, err := rules.NewWAF(version, conf.GetFlag().GetRulePath(), flag)
	if err != nil {
		return err
	}

	w.wafMap.Store(namespace, waf)
	if waf != nil {
		zlog.Infof("[edge] WAF initialized successfully, ns=%s", namespace)
	}

	return nil
}

func (w *WAFCache) Load(namespace string) coraza.WAF {
	value, ok := w.wafMap.Load(namespace)
	if !ok {
		return nil
	}

	return value
}

func (w *WAFCache) LoadContext(ctx *httpxhz.Context) coraza.WAF {
	hCtx, ok := httpxhz.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge] hit cached rules of namespace %s", hCtx.GetTenantNs())
	return w.Load(hCtx.GetTenantNs())
}
