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

// Package routing provides the WAF handler.
package routing

import (
	"net/http"

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
)

func NewMockRouter() chains.Handler {
	return &Mock{
		BaseHandler: chains.New(),
	}
}

type Mock struct {
	*chains.BaseHandler
}

func (r *Mock) Handle(ctx corehttp.Context) error {
	return ctx.String(http.StatusOK, "")
}
