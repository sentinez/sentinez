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

package httpxdmz

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez"
	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/common/uuidx"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
)

func WrapHandler(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {

		requestId := uuidx.NewNanoID(sentinez.PrefixRequestID)
		c.Request.Header.Set(corehttp.HeaderXRequest, requestId)

		ctx = httpxcmn.SetRequestTime(ctx)

		next(ctx, c)
	}
}
