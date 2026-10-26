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

package szhttp

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sentinez/core/internal/render"
	"github.com/sentinez/core/szbytestr"
)

// ForbiddenPage returns the HTML 403 page for the request requestId.
func ForbiddenPage(requestId string) templ.Component {
	return render.Forbidden(requestId)
}

func Forbidden(ctx Context) error {
	return renderPage(ctx, http.StatusForbidden, render.Forbidden,
		szbytestr.AccessDenied)
}

func BadRequest(ctx Context) error {
	return renderPage(ctx, http.StatusBadRequest, render.BadRequest,
		szbytestr.BadRequest)
}

func InternalServerError(ctx Context) error {
	return renderPage(ctx, http.StatusInternalServerError,
		render.InternalError, szbytestr.InternalServerError)
}

func NotFound(ctx Context) error {
	return renderPage(ctx, http.StatusNotFound, render.NotFound,
		szbytestr.NotFound)
}

func TooManyRequests(ctx Context) error {
	return renderPage(ctx, http.StatusTooManyRequests,
		render.TooManyRequests, szbytestr.TooManyRequests)
}

// Challenge writes the proof-of-work page: the client solves challenge at
// difficulty (see szcrypto/pow) and posts the nonce to action.
func Challenge(
	ctx Context, challenge string, difficulty int, action string,
) error {
	ctx.SetResponseHeader(szbytestr.HeaderCacheControl, szbytestr.ValueNoStore)

	page := func(requestId string) templ.Component {
		return render.Challenge(requestId, challenge, difficulty, action)
	}

	return renderPage(ctx, http.StatusForbidden, page, szbytestr.AccessDenied)
}

// renderPage writes the HTML error page, or fallback as plain text when the
// page cannot be rendered.
func renderPage(
	ctx Context,
	statusCode int,
	page func(requestId string) templ.Component,
	fallback []byte,
) error {
	err := ctx.Render(statusCode, page(ctx.RequestId()))
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(statusCode, fallback)
	}

	return nil
}
