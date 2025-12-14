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
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	once    sync.Once
	wafInst *WAFCache
	mu      sync.Mutex
)

func New() *WAFCache {
	once.Do(func() {
		wafInst = &WAFCache{
			space: ssync.NewMap[string, coraza.WAF](),
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
	space *ssync.Map[string, coraza.WAF]
}

func (w *WAFCache) Store(conf *confpb.Config, namespace string,
	version corers.Version, flag corers.Flag) error {

	if w == nil {
		return nil
	}

	waf, err := corers.NewWAF(version, conf.GetFlag().GetRulePath(), flag)
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

func (w *WAFCache) LoadContext(ctx corehttp.Context) coraza.WAF {
	if w == nil {
		return nil
	}

	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok {
		return nil
	}

	zlog.Debugf("[edge][namespace] hit waf cached %s", hCtx.GetTenantNs())
	return w.Load(hCtx.GetTenantNs())
}

func Store(conf *confpb.Config,
	namespace string, version corers.Version, flag corers.Flag) error {
	mu.Lock()
	defer mu.Unlock()

	if wafInst == nil {
		wafInst = New()
	}

	return wafInst.Store(conf, namespace, version, flag)
}
