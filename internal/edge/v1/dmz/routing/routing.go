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
	"fmt"
	"net/http"
	"strings"
	"sync"

	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	"github.com/sentinez/sentinez/pkg/core/net/httpx/hz/proxy"
	"github.com/sentinez/sentinez/pkg/std/stderr"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	"github.com/sentinez/sentinez/pkg/syncx"
)

var (
	dynamic   *syncx.Map[string, string]
	rewrite   *syncx.Map[string, string]
	proxyInst *proxy.ReverseProxy
	once      sync.Once
)

type Router struct{}

func init() {
	once.Do(func() {
		dynamic = syncx.NewMap[string, string]()
		rewrite = syncx.NewMap[string, string]()
	})
}

func key(ns, prefix string) string {
	return fmt.Sprintf("%s|%s", ns, prefix)

}

func (r *Router) Store(
	proxy *proxy.ReverseProxy, config *edgeyaml.Config) *Router {

	proxyInst = proxy

	for _, proxy := range config.ReverseProxies {
		for _, routeConfig := range proxy.Routes {
			zlog.Debugf(
				"[edge] routing store: ns=%s prefix=%s -> %s (rewrite: %s)",
				proxy.Namespace, routeConfig.MatchPrefix,
				routeConfig.Target, routeConfig.Rewrite,
			)

			target, ok := dynamic.Load(routeConfig.MatchPrefix)
			if ok && target != "" {
				zlog.Warnf(
					"[edge] duplicate prefix: %s -> %s (new: %s), ignoring",
					routeConfig.MatchPrefix, target, routeConfig.Target,
				)

				continue
			}

			if routeConfig.Rewrite == "" {
				routeConfig.Rewrite = routeConfig.MatchPrefix
			}

			k := key(proxy.Namespace, routeConfig.MatchPrefix)

			dynamic.Store(k, routeConfig.Target)
			rewrite.Store(k, routeConfig.Rewrite)
		}
	}

	return &Router{}
}

func Serve(edgeYml *edgeyaml.Config, server httpxhz.Server) error {

	reverseProxy, err := proxy.NewReverseProxy()
	if err != nil {
		zlog.Errorf("failed to create proxy instance: %v", err)
		return err
	}

	r := &Router{}

	handler := r.Store(reverseProxy, edgeYml).Match()

	server.Handle(handler)

	return nil
}

func (r *Router) Match() func(ctx *httpxhz.Context) error {
	return func(ctx *httpxhz.Context) error {
		zlog.Debugf("[edge] request host: %s", string(ctx.Host()))

		if proxyInst == nil {
			return ctx.String(http.StatusInternalServerError,
				"proxy not initialized")
		}

		target, err := match(ctx)
		if err != nil {
			zlog.Error("[edge] routing match error: ", err)
			return ctx.String(http.StatusNotFound, "not found")
		}

		proxyInst.Serve(ctx, target)

		return nil
	}
}

func match(ctx *httpxhz.Context) (string, error) {
	hCtx, ok := httpxhz.GetRequestContext(ctx)
	if !ok || hCtx.GetTenantNs() == "" {
		return "", stderr.F("unknown namespace of request")
	}

	origin := ""
	path := ctx.Path()
	matchPrefix := prefixPath(path)
	k := key(hCtx.GetTenantNs(), matchPrefix)
	defaultKey := key(hCtx.TenantNs, "/")

	rewritePrefix, ok := rewrite.Load(k)
	if !ok || rewritePrefix == "" {
		rewritePrefix = matchPrefix
	}

	target, ok := dynamic.Load(k)
	if ok {
		zlog.Debugf(
			"[edge] routing match: ns=%s prefix=%s -> %s (prefix: %s)",
			hCtx.GetTenantNs(), rewritePrefix, target, matchPrefix,
		)

		remainingPath := strings.TrimPrefix(path, matchPrefix) + rewritePrefix
		ctx.Request.URI().SetPath(remainingPath)
		origin = target

		return target, nil
	}

	if origin, ok = dynamic.Load(defaultKey); ok {
		return origin, nil
	}

	return "", stderr.F("not found: %s", path)
}

func prefixPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) > 0 && parts[0] != "" {
		return "/" + parts[0]
	}
	return "/"
}
