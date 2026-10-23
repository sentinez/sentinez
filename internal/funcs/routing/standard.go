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
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/chains"
	"github.com/sentinez/sentinez/internal/envplugin"
	"github.com/sentinez/sentinez/internal/memory/reverseproxy"
	"github.com/sentinez/sentinez/internal/memory/routes"
	"github.com/sentinez/shared/szlog"
)

// StandardRouterId is the chain stage name of StandardRouter.
const StandardRouterId = "ROU"

var (
	_ chains.Plugin                      = (*StandardRouter)(nil)
	_ chains.Provisioner[*envplugin.Env] = (*StandardRouter)(nil)
)

// StandardRouter proxies the request to the upstream of its route. It
// ends the chain: next is never called.
type StandardRouter struct {
	route *routes.Router
	proxy *reverseproxy.ReverseProxy
}

func (*StandardRouter) ID() string { return StandardRouterId }

func (r *StandardRouter) Provision(env *envplugin.Env) error {
	r.route = env.Store().Route()
	r.proxy = env.Store().ReverseProxy()
	return nil
}

func (r *StandardRouter) Handle(ctx szhttp.Context, _ chains.Handler) error {
	target, err := r.route.Match(ctx)
	if err != nil {
		szlog.Error("[edge] routing match error: ", err)
		return szhttp.NotFound(ctx)
	}

	rprx, ok := r.proxy.Load(target)
	if ok {
		rprx.Serve(ctx)
	}

	return nil
}
