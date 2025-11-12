// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secure

import (
	corehttp "github.com/sentinez/sentinez/core/http"
	corerules "github.com/sentinez/sentinez/core/rules"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
)

type Rule struct {
	chains.BaseHandler
	ingress corerules.Rules
}

func (r *Rule) Handler(ctx corehttp.Context) error {

	if ok := r.ingress.Eval(ctx, nil); ok {
		return httpxcmn.Forbidden(ctx)
	}

	return r.HandleNext(ctx)
}
