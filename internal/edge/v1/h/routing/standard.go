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

package routing

import (
	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	"github.com/sentinez/sentinez/pkg/dmz/mem/routes"
	stdproxy "github.com/sentinez/sentinez/pkg/network/httpx/std/proxy"
	"github.com/sentinez/shared/zlog"
)

func NewStandardRouter() chains.Handler {
	reverseProxy, err := stdproxy.NewReverseProxy()
	if err != nil {
		zlog.Errorf("failed to create proxy instance: %v", err)
	}

	return &StandardRouter{
		BaseHandler: chains.New(),
		handler:     routes.GetRouter().SetReverseProxy(reverseProxy),
	}
}

type StandardRouter struct {
	*chains.BaseHandler
	handler func(ctx corehttp.Context) error
}

func (r *StandardRouter) Handle(ctx corehttp.Context) error {
	// zlog.Debug("[edge] >>> visit standard router")

	return r.handler(ctx)
}
