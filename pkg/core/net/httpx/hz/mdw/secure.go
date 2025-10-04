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

package httpxhzmdw

import (
	"github.com/corazawaf/coraza/v3/types"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	httpxhzsec "github.com/sentinez/sentinez/pkg/core/net/httpx/hz/sec"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/sentinez/sentinez/rules"
)

func ProtectedWithCallback(
	ruleBasePath string,
	cb func(*httpxhz.Context, types.Transaction),
) func(httpxhz.RequestHandler) httpxhz.RequestHandler {

	flag := rules.ReqAppAttackRCE

	waf, err := rules.NewWAF(rules.Ver4_16_0, ruleBasePath, flag)
	if err != nil {
		zlog.Errorf("[httpxhzsec] failed to create WAF: %v", err)
		return nil
	}

	if waf != nil {
		zlog.Info("[httpxhzsec] WAF initialized successfully")
	}

	return func(next httpxhz.RequestHandler) httpxhz.RequestHandler {
		return httpxhzsec.WrapHandlerWithCallback(waf, next, cb)
	}
}
