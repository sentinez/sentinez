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
	"bytes"
	"sort"
	"sync"

	"github.com/sentinez/core/szbytestr"
	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/core/szhttp/variable"
	edgepb "github.com/sentinez/sentinez/api/proto/sentinez/dmz/edge/v1"
	"github.com/sentinez/shared/bytesconv"
	"github.com/sentinez/shared/szerror"
	"github.com/sentinez/shared/szlog"
	"github.com/sentinez/shared/szsync"
)

var (
	_inst   *Router
	_once   sync.Once
	_buffer = szsync.NewPool[bytes.Buffer]()
)

func New() *Router {
	_once.Do(func() {
		_inst = &Router{
			routes: szsync.NewMap[string, []*edgepb.Location](),
		}
	})

	return _inst
}

type Router struct {
	routes *szsync.Map[string, []*edgepb.Location]
}

// nolint:funlen
func (r *Router) Store(server *edgepb.Server) {
	if len(server.GetLocations()) == 0 {
		return
	}

	var validRoutes []*edgepb.Location
	for _, routeConfig := range server.GetLocations() {
		szlog.Debugf(
			"edge: ns=%s %s -> %s (rewrite: %s)",
			server.GetName(), routeConfig.GetLocation(),
			routeConfig.GetProxyPass(), routeConfig.GetProxyRewrite(),
		)

		route := &edgepb.Location{
			Location:        routeConfig.GetLocation(),
			ProxyPass:       routeConfig.GetProxyPass(),
			ProxyRewrite:    routeConfig.GetProxyRewrite(),
			ProxySetHeaders: make(map[string]string),
		}

		for k, v := range routeConfig.GetProxySetHeaders() {
			if variable.IsValidHeaderKey(k) {
				route.ProxySetHeaders[k] = v
			} else {
				szlog.Warnf(
					"invalid proxy header %q in ns=%q location=%q, ignoring",
					k, server.GetName(), routeConfig.GetLocation(),
				)
			}
		}

		if route.GetProxyRewrite() == "" {
			route.ProxyRewrite = route.GetLocation()
		}

		validRoutes = append(validRoutes, route)
	}

	// sort routes descending by location
	// length for longest-prefix match (like nginx)
	sort.Slice(validRoutes, func(i, j int) bool {
		return len(validRoutes[i].GetLocation()) >
			len(validRoutes[j].GetLocation())
	})

	r.routes.Store(server.GetName(), validRoutes)
}

// nolint:funlen
// Match picks the route of the request and rewrites it for the upstream:
// the location a SecRule ROUTE_TO action chose (ContextExtra.route_location)
// when the namespace has it, else the longest location prefixing the path.
func (r *Router) Match(ctx szhttp.Context) (string, error) {
	namespace := ctx.X().GetNamespace()
	path := ctx.Path()

	routes, ok := r.routes.Load(namespace)
	if !ok {
		return "", szerror.F("not found: %s", path)
	}

	route := pick(routes, ctx.X().GetRouteLocation(), path)
	if route == nil {
		return "", szerror.F("not found: %s", path)
	}

	szlog.Debugf(
		"edge: routing match: ns=%s prefix=%s -> %s (prefix: %s)",
		namespace, route.GetProxyRewrite(), route.GetProxyPass(),
		route.GetLocation(),
	)

	return forward(ctx, route)
}

// pick returns the route named location, else the longest one prefixing
// path (routes are sorted longest first), or nil.
func pick(
	routes []*edgepb.Location, location string, path []byte,
) *edgepb.Location {
	if location != "" {
		for _, route := range routes {
			if route.GetLocation() == location {
				return route
			}
		}
	}

	for _, route := range routes {
		if matchLocation(path, bytesconv.S2b(route.GetLocation())) {
			return route
		}
	}

	return nil
}

// forward rewrites the request path and headers for route and returns its
// upstream server.
func forward(ctx szhttp.Context, route *edgepb.Location) (string, error) {
	locationBytes := bytesconv.S2b(route.GetLocation())
	rewriteBytes := bytesconv.S2b(route.GetProxyRewrite())

	// path = /api/v1/users, location = /api, rewrite = /v1
	// remainingPath = /v1 + /v1/users = /v1/v1/users
	remainingPath := _buffer.Get()
	defer func() {
		remainingPath.Reset()
		_buffer.Put(remainingPath)
	}()

	joinRewrite(remainingPath, rewriteBytes,
		bytes.TrimPrefix(ctx.Path(), locationBytes))

	ctx.SetPath(remainingPath.Bytes())

	// Set default proxy routing headers
	ctx.SetHeader(szbytestr.HeaderXForwardedPrefix, locationBytes)

	// Add custom headers from config
	for hk, hv := range route.GetProxySetHeaders() {
		val, err := variable.ParseProxyVal(ctx, hv)
		if err != nil {
			return "", err
		}

		ctx.SetHeader(bytesconv.S2b(hk), bytesconv.S2b(val))
	}

	return route.GetProxyPass()[0].GetServer(), nil
}

func matchLocation(path, location []byte) bool {
	if !bytes.HasPrefix(path, location) {
		return false
	}

	return len(path) == len(location) ||
		bytes.HasSuffix(location, []byte("/")) ||
		path[len(location)] == '/'
}

// joinRewrite writes rewrite+rest without doubling the slash between them,
// e.g. rewrite "/" and rest "/x" give "/x" rather than "//x".
func joinRewrite(dst *bytes.Buffer, rewrite, rest []byte) {
	if bytes.HasSuffix(rewrite, []byte("/")) &&
		bytes.HasPrefix(rest, []byte("/")) {
		rest = rest[1:]
	}

	dst.Write(rewrite)
	dst.Write(rest)
}
