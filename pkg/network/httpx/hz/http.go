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

package httpxhz

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/sentinez/sentinez/pkg/templx"
)

func WrapHandler(next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ctx = setIdentifier(ctx)
		next(ctx, c)
	}
}

func Forbidden(ctx *Context) {

	ctx.SetStatusCode(http.StatusForbidden)
	ctx.Response.Header.Set("Content-Type", "text/plain; charset=utf-8")

	ctx.SetContentType("text/html; charset=utf-8")
	err := templx.Forbidden().Render(ctx.Context(), ctx.Response.BodyWriter())
	if err != nil {
		ctx.SetBodyString("Access denied")
	}
}
