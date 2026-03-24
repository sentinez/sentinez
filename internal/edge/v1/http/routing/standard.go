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
	"github.com/sentinez/sentinez/internal/shared/mem/reverseproxy"
	"github.com/sentinez/sentinez/internal/shared/mem/routes"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	"github.com/sentinez/shared/zlog"
)

func NewStandardRouter() chains.Handler {
	return &StandardRouter{
		BaseHandler:  chains.New(),
		router:       routes.GetRouter(),
		reverseProxy: reverseproxy.GetEngine(),
	}
}

type StandardRouter struct {
	*chains.BaseHandler
	reverseProxy *reverseproxy.ReverseProxy
	router       *routes.Router
}

func (r *StandardRouter) Handle(ctx corehttp.Context) error {
	if r.reverseProxy == nil {
		zlog.Error("[edge][routing]: proxy not initialized")
		return httpxcmn.InternalServerError(ctx)
	}

	target, err := r.router.Match(ctx)
	if err != nil {
		zlog.Error("[edge] routing match error: ", err)
		return httpxcmn.NotFound(ctx)
	}

	r.reverseProxy.Load(target).Serve(ctx)

	return nil
}
