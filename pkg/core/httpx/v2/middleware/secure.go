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

package httpv2mdw

import (
	"github.com/sentinez/sentinez/pkg/core/dmz/secure"
	http2sec "github.com/sentinez/sentinez/pkg/core/dmz/secure/http2"
	"github.com/valyala/fasthttp"
)

func Protected(ruleRoot string,
) func(fasthttp.RequestHandler) fasthttp.RequestHandler {

	waf := secure.NewFireWall(ruleRoot)

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return http2sec.WrapHandler(waf, next)
	}
}
