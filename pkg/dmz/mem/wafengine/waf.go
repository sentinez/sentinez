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
	configspb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/configs/v1"
	"github.com/sentinez/sentinez/core/rulesets"
	"github.com/sentinez/sentinez/pkg/common/syncx"
	httpxdmz "github.com/sentinez/sentinez/pkg/network/httpx/dmz"
	"github.com/sentinez/sentinez/pkg/zlog"
)

var (
	once    sync.Once
	wafInst *WAFCache
)

func New() *WAFCache {
	once.Do(func() {
		wafInst = &WAFCache{
			space: syncx.NewMap[string, coraza.WAF](),
		}
	})

	return wafInst
}

func GetEngine() *WAFCache {
	return wafInst
}

type WAFCache struct {
	// key: namespace
	// ex: dev.sentinez.vn
	//	- domain: sentinez.vn
	// 	- namespace: dev
	space *syncx.Map[string, coraza.WAF]
}

func (w *WAFCache) Store(conf *configspb.AppConfig, namespace string,
	version rulesets.Version, flag rulesets.Flag) error {

	if w == nil {
		return nil
	}

	waf, err := rulesets.NewWAF(version, conf.GetFlag().GetRulePath(), flag)
	if err != nil {
		return err
	}

	w.space.Store(namespace, waf)
	if waf != nil {
		zlog.Infof("[edge] WAF initialized successfully, ns=%s", namespace)
	}

	return nil
}

func (w *WAFCache) Load(namespace string) coraza.WAF {
	if w == nil {
		return nil
	}

	value, ok := w.space.Load(namespace)
	if !ok {
		return nil
	}

	return value
}

func (w *WAFCache) LoadContext(ctx *httpxdmz.Context) coraza.WAF {
	if w == nil {
		return nil
	}

	hCtx, ok := httpxdmz.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge][namespace] hit waf cached %s", hCtx.GetTenantNs())
	return w.Load(hCtx.GetTenantNs())
}
