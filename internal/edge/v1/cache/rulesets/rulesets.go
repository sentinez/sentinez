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

package rulesets

import (
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/std/common/v1"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/sentinez/sentinez/pkg/syncx"
	"github.com/sentinez/sentinez/rules"
)

var (
	// key: namespace
	// ex: dev.sentinez.vn
	//	- domain: sentinez.vn
	// 	- namespace: dev
	rulemap *syncx.Map[string, coraza.WAF]

	once sync.Once
)

func init() {
	once.Do(func() {
		rulemap = syncx.NewMap[string, coraza.WAF]()
	})
}

func Store(appConf *common.AppConfig,
	namespace string, rulesetsFlag rules.RulesetsFlag) error {

	waf, err := rules.NewWAF(rules.Ver4_16_0,
		appConf.GetFlag().GetRulePath(), rulesetsFlag)
	if err != nil {
		return err
	}

	rulemap.Store(namespace, waf)
	if waf != nil {
		zlog.Infof("[edge] WAF initialized successfully, ns=%s", namespace)
	}

	return nil
}

func Load(namespace string) coraza.WAF {
	value, ok := rulemap.Load(namespace)
	if !ok {
		return nil
	}

	return value
}
