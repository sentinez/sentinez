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
	"sync"

	"github.com/a-h/templ"
	"github.com/sentinez/core/internal/render"
	"github.com/sentinez/core/szbytestr"
)

var _mu sync.Mutex

// ForbiddenPage returns the HTML 403 page for the request requestId.
func ForbiddenPage(requestId string) templ.Component {
	return render.Forbidden(requestId)
}

func Forbidden(ctx Context) error {
	_mu.Lock()
	defer _mu.Unlock()

	err := ctx.Render(http.StatusForbidden, render.Forbidden(ctx.RequestId()))
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(http.StatusForbidden, szbytestr.AccessDenied)
	}

	return nil
}

func BadRequest(ctx Context) error {
	_mu.Lock()
	defer _mu.Unlock()

	err := ctx.Render(
		http.StatusBadRequest, render.BadRequest(ctx.RequestId()),
	)
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(http.StatusBadRequest, szbytestr.BadRequest)
	}

	return nil
}

func InternalServerError(ctx Context) error {
	_mu.Lock()
	defer _mu.Unlock()

	err := ctx.Render(
		http.StatusInternalServerError, render.InternalError(ctx.RequestId()))
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(
			http.StatusInternalServerError, szbytestr.InternalServerError)
	}

	return nil
}

func NotFound(ctx Context) error {
	_mu.Lock()
	defer _mu.Unlock()

	err := ctx.Render(http.StatusNotFound, render.NotFound(ctx.RequestId()))
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(http.StatusNotFound, szbytestr.NotFound)
	}

	return nil
}

func TooManyRequests(ctx Context) error {
	_mu.Lock()
	defer _mu.Unlock()

	err := ctx.Render(http.StatusTooManyRequests,
		render.TooManyRequests(ctx.RequestId()))
	if err != nil {
		ctx.ResetResponse()
		return ctx.String(http.StatusTooManyRequests, szbytestr.TooManyRequests)
	}

	return nil
}
