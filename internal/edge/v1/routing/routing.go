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
	"strings"
	"sync"

	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1/proxy"
	syncx "github.com/sentinez/sentinez/pkg/common/sync"
	httpxv2 "github.com/sentinez/sentinez/pkg/core/httpx/v2"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

var (
	dynamic   *syncx.Map[string, string]
	rewrite   *syncx.Map[string, string]
	proxyInst *proxy.Proxy
	once      sync.Once
)

func init() {
	once.Do(func() {
		dynamic = syncx.NewMap[string, string]()
		rewrite = syncx.NewMap[string, string]()
	})
}

func Store(proxy *proxy.Proxy, config *edgeyaml.Config) {
	proxyInst = proxy

	for _, routeConfig := range config.Proxy.Routes {
		zlog.Debugf(
			"[edge] routing store: %s -> %s (rewrite: %s)",
			routeConfig.MatchPrefix, routeConfig.Target, routeConfig.Rewrite,
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

		dynamic.Store(routeConfig.MatchPrefix, routeConfig.Target)
		rewrite.Store(routeConfig.MatchPrefix, routeConfig.Rewrite)
	}
}

func Match() func(ctx *httpxv2.Context) error {
	return func(ctx *httpxv2.Context) error {
		if proxyInst == nil {
			return ctx.String(http.StatusInternalServerError,
				"proxy not initialized")
		}

		target, err := match(ctx)
		if err != nil {
			zlog.Error("[edge] routing match error: ", err)
			return ctx.String(http.StatusNotFound, "not found")
		}

		if err := proxyInst.ServeHTTP(ctx, target); err != nil {
			zlog.Error("[edge] routing proxy error: ", err)
			return ctx.String(http.StatusInternalServerError, err.Error())
		}

		return nil
	}
}

func match(ctx *httpxv2.Context) (string, error) {
	path := ctx.Path()
	matchPrefix := prefixPath(path)
	origin := ""

	rewritePrefix, ok := rewrite.Load(matchPrefix)
	if !ok || rewritePrefix == "" {
		rewritePrefix = matchPrefix
	}

	target, ok := dynamic.Load(matchPrefix)
	if ok {
		zlog.Debugf(
			"[edge] routing match: %s -> %s (prefix: %s)",
			rewritePrefix, target, matchPrefix,
		)

		remainingPath := strings.TrimPrefix(path, matchPrefix) + rewritePrefix
		ctx.Request.URI().SetPath(remainingPath)
		origin = target

		return target, nil
	}

	if origin == "" {
		if origin, ok = dynamic.Load("/"); ok {
			return origin, nil
		}
	}

	return "", errors.F("not found: %s", path)
}

func prefixPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) > 0 && parts[0] != "" {
		return "/" + parts[0]
	}
	return "/"
}
