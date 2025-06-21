// Copyright 2025 Sentinez Labs.
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

package httpxsecure

import (
	"os"
	"sync"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	waf  coraza.WAF
	lock sync.Mutex
)

func NewFireWall(confPath string, ruleRoot string) coraza.WAF {
	lock.Lock()
	defer lock.Unlock()

	if waf == nil {
		var err error

		rootFS := os.DirFS(ruleRoot)
		conf := coraza.NewWAFConfig().WithRootFS(rootFS).
			WithErrorCallback(logError).
			WithDirectivesFromFile(confPath)

		waf, err = coraza.NewWAF(conf)
		if err != nil {
			zlog.Errorf("failed to create WAF: %v", err)
			return nil
		}
	}

	return waf
}

func logError(err types.MatchedRule) {
	msg := err.ErrorLog()
	zlog.Debugf("[%s] %s", err.Rule().Severity(), msg)
}
