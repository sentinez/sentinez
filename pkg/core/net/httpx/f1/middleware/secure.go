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

package httpxf1mdw

import (
	"github.com/corazawaf/coraza/v3/types"
	httpxf1 "github.com/sentinez/sentinez/pkg/core/net/httpx/f1"
	"github.com/sentinez/sentinez/pkg/core/net/secure"
	httpfsec "github.com/sentinez/sentinez/pkg/core/net/secure/httpf"
)

func ProtectedWithCallback(
	ruleBasePath string,
	cb func(*httpxf1.Context, types.Transaction),
) func(httpxf1.RequestHandler) httpxf1.RequestHandler {

	waf := secure.NewFireWall(ruleBasePath)

	return func(next httpxf1.RequestHandler) httpxf1.RequestHandler {
		return httpfsec.WrapHandlerWithCallback(waf, next, cb)
	}
}
