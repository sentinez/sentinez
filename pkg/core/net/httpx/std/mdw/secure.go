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

package httpv1mdw

import (
	"net/http"

	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/sentinez/sentinez/pkg/zlog"
	"github.com/sentinez/sentinez/security"
	"github.com/sentinez/sentinez/security/wafengine"
)

func Protected(ruleBasePath string) func(next http.Handler) http.Handler {

	flag := security.ReqAppAttackRCE

	waf, err := wafengine.NewWAF(security.CRSv4160, ruleBasePath, flag)
	if err != nil {
		zlog.Errorf("[httpxf1] failed to create WAF: %v", err)
		return nil
	}

	if waf != nil {
		zlog.Info("[httpxf1] WAF initialized successfully")
	}

	return func(next http.Handler) http.Handler {
		return txhttp.WrapHandler(waf, next)
	}
}
