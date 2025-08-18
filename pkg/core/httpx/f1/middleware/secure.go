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
	"github.com/sentinez/sentinez/pkg/core/defense/secure"
	httpfsec "github.com/sentinez/sentinez/pkg/core/defense/secure/httpf"
	"github.com/valyala/fasthttp"
)

func Protected(ruleBasePath string,
) func(fasthttp.RequestHandler) fasthttp.RequestHandler {

	waf := secure.NewFireWall(ruleBasePath)

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return httpfsec.WrapHandler(waf, next)
	}
}
