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
	"fmt"
	"io/fs"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szrs"
	"github.com/sentinez/sentinez"
	corerulesetpb "github.com/sentinez/sentinez/api/proto/sentinez/types/coreruleset/v1"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	_once         sync.Once
	_rulesetsInst *Rulesets
)

func New() *Rulesets {
	_once.Do(func() {
		_rulesetsInst = &Rulesets{
			space: ssync.NewMap[string, coraza.WAF](),
		}
	})

	return _rulesetsInst
}

type Rulesets struct {
	// key: namespace
	// ex: dev.sentinez.vn
	//	- domain: sentinez.vn
	// 	- namespace: dev
	space *ssync.Map[string, coraza.WAF]
}

func (rs *Rulesets) Store(
	namespace string, setting *corerulesetpb.CoreRuleset,
) error {

	if rs == nil {
		return nil
	}

	waf, err := NewWAF(setting)
	if err != nil {
		return err
	}

	rs.space.Store(namespace, waf)
	if waf != nil {
		zlog.Infof("edge: WAF initialized successfully, ns=%s", namespace)
	}

	return nil
}

func (rs *Rulesets) Load(namespace string) (coraza.WAF, bool) {
	if rs == nil {
		return nil, false
	}

	value, ok := rs.space.Load(namespace)
	if !ok {
		return nil, false
	}

	return value, true
}

func (rs *Rulesets) LoadContext(ctx szhttp.Context) (coraza.WAF, bool) {
	if rs == nil {
		return nil, false
	}

	zlog.Debugf("edge: namespace: hit WAF cached %s", ctx.X().GetNamespace())
	return rs.Load(ctx.X().GetNamespace())
}

// NewWAF builds the coraza.WAF of setting from the embedded CRS of its
// version.
func NewWAF(setting *corerulesetpb.CoreRuleset) (coraza.WAF, error) {
	rootFs, err := crsMap(setting.GetVersion())
	if err != nil {
		return nil, err
	}

	return szrs.NewWAF(setting.GetVersion(), rootFs, setting)
}

func crsMap(version corerulesetpb.Version) (fs.FS, error) {
	switch version {
	case corerulesetpb.Version_VERSION_V4_16_0:
		return sentinez.WAF4160(), nil
	case corerulesetpb.Version_VERSION_V4_17_0:
		return sentinez.WAF4170(), nil
	}

	return nil, fmt.Errorf("waf: coreruleset version invalid")
}
