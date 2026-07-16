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

package corerule

import (
	corehttp "github.com/sentinez/core/http"
	rulepb "github.com/sentinez/sentinez/api/gen/go/sentinez/secure/rule/v1"
)

func visit(ctx corehttp.RequestContext, cond *rulepb.Condition) bool {
	// zlog.Debugf("ev: visit with source: %s", cond.GetSource())

	switch cond.GetSource() {

	case rulepb.FieldSource_FIELD_SOURCE_PATH:
		return matchSourcePath(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_QUERY:
		return matchSourceQuery(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_BODY:
		return matchSourceBody(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_HEADER:
		return matchSourceHeader(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_METHOD:
		return matchSourceMethod(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_HOST:
		return matchSourceHost(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_IP:
		return matchSourceIP(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_TLS:
		return matchSourceTLS(ctx, cond)

	case rulepb.FieldSource_FIELD_SOURCE_JA4:
		return matchSourceJA4(ctx, cond)

	default:
		return bypass
	}
}
