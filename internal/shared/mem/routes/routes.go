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

package routes

import (
	"sort"
	"strings"
	"sync"

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/core/http/variable"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/pkg/common/errorx"
	ssync "github.com/sentinez/shared/sync"
	"github.com/sentinez/shared/zlog"
)

var (
	inst *Router
	once sync.Once
	mu   sync.Mutex
)

func NewRouter() *Router {
	once.Do(func() {
		inst = &Router{
			routes: ssync.NewMap[string, []*edgepb.Location](),
		}
	})

	return inst
}

func GetRouter() *Router {
	return inst
}

type Router struct {
	routes *ssync.Map[string, []*edgepb.Location]
}

// nolint:funlen
func (r *Router) Store(server *edgepb.Server) {
	if len(server.Locations) == 0 {
		return
	}

	var validRoutes []*edgepb.Location
	for _, routeConfig := range server.Locations {
		zlog.Debugf(
			"[edge] ns=%s %s -> %s (rewrite: %s)",
			server.Name, routeConfig.Location,
			routeConfig.ProxyPass, routeConfig.ProxyRewrite,
		)

		route := &edgepb.Location{
			Location:        routeConfig.Location,
			ProxyPass:       routeConfig.ProxyPass,
			ProxyRewrite:    routeConfig.ProxyRewrite,
			ProxySetHeaders: make(map[string]string),
		}

		for k, v := range routeConfig.ProxySetHeaders {
			if variable.IsValidHeaderKey(k) {
				route.ProxySetHeaders[k] = v
			} else {
				zlog.Warnf(
					"invalid proxy header %q in ns=%q location=%q, ignoring",
					k, server.Name, routeConfig.Location,
				)
			}
		}

		if route.ProxyRewrite == "" {
			route.ProxyRewrite = route.Location
		}

		validRoutes = append(validRoutes, route)
	}

	// sort routes descending by location
	// length for longest-prefix match (like nginx)
	sort.Slice(validRoutes, func(i, j int) bool {
		return len(validRoutes[i].Location) > len(validRoutes[j].Location)
	})

	r.routes.Store(server.Name, validRoutes)
}

// nolint:funlen
func (r *Router) Match(ctx corehttp.Context) (string, error) {
	hCtx, ok := corehttp.GetRequestContext(ctx)
	if !ok || hCtx.GetServerName() == "" {
		return "", errorx.F("unknown server name of request")
	}

	serverName := hCtx.GetServerName()
	path := ctx.Path()

	routes, ok := r.routes.Load(serverName)
	if !ok {
		return "", errorx.F("not found: %s", path)
	}

	for _, route := range routes {
		if strings.HasPrefix(path, route.Location) {
			zlog.Debugf(
				"[edge] routing match: ns=%s prefix=%s -> %s (prefix: %s)",
				serverName, route.ProxyRewrite, route.ProxyPass, route.Location,
			)

			// path = /api/v1/users, location = /api, rewrite = /v1
			// remainingPath = /v1 + /v1/users = /v1/v1/users
			remainingPath := route.ProxyRewrite +
				strings.TrimPrefix(path, route.Location)
			ctx.SetPath(remainingPath)

			// Set default proxy routing headers
			ctx.SetHeader("X-Forwarded-Prefix", route.Location)

			// Add custom headers from config
			for hk, hv := range route.ProxySetHeaders {
				val, err := variable.ParseProxyVal(ctx, hv)
				if err != nil {
					return "", err
				}

				ctx.SetHeader(hk, val)
			}

			return route.GetProxyPass()[0].GetServer(), nil
		}
	}

	return "", errorx.F("not found: %s", path)
}

func Store(server *edgepb.Server) {
	mu.Lock()
	defer mu.Unlock()

	if inst == nil {
		inst = NewRouter()
	}

	inst.Store(server)
}
