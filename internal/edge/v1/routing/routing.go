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
	static    *syncx.Map[string, string]
	proxyInst *proxy.Proxy
	once      sync.Once
)

func init() {
	once.Do(func() {
		dynamic = syncx.NewMap[string, string]()
		static = syncx.NewMap[string, string]()
	})
}

func Store(proxy *proxy.Proxy, config *edgeyaml.Routes) {
	proxyInst = proxy

	for _, routeConfig := range config.Routes {
		// Store the path and target in the sync.Map
		if routeConfig.Static {
			static.Store(routeConfig.Location, routeConfig.ProxyPass)
			continue
		}

		dynamic.Store(routeConfig.Location, routeConfig.ProxyPass)
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
			zlog.Debug("[edge] routing match error: ", err)
			return ctx.String(http.StatusNotFound, "not found")
		}

		return proxyInst.ServeHTTP(ctx, target)
	}
}

func match(ctx *httpxv2.Context) (string, error) {
	pathRequest := ctx.Path()
	targetRequest := ""

	location := firstPrefix(pathRequest)
	proxyPass, ok := dynamic.Load(location)
	if ok {
		targetRequest = proxyPass
		remainingPath := strings.TrimPrefix(pathRequest, location)
		ctx.Request.URI().SetPath(remainingPath)
	}

	if !ok {
		proxyPass, ok = static.Load(location)
		if ok {
			targetRequest = proxyPass
		}
	}

	if targetRequest == "" {
		targetRequest, ok = dynamic.Load("/")
		if !ok {
			return "", errors.F("not found: %s", pathRequest)
		}
	}

	return targetRequest, nil
}

func firstPrefix(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.SplitN(path, "/", 2)

	if len(parts) > 0 && parts[0] != "" {
		return "/" + parts[0]
	}
	return "/"
}
