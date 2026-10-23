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

// Package routing provides the plugins ending the edge chain.
package routing

import (
	"net/http"

	"github.com/sentinez/core/szhttp/chains"

	"github.com/sentinez/core/szhttp"
)

// MockId is the chain stage name of Mock.
const MockId = "MCK"

var _ chains.Plugin = (*Mock)(nil)

// Mock answers every request with an empty 200.
type Mock struct{}

func (*Mock) ID() string { return MockId }

func (*Mock) Handle(ctx szhttp.Context, _ chains.Handler) error {
	return ctx.String(http.StatusOK, []byte(""))
}
