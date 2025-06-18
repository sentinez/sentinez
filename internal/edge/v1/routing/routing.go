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
	"strings"
	"sync"

	edgeyaml "github.com/sentinez/sentinez/cmd/edge/v1/apps/yaml"
	"github.com/sentinez/sentinez/internal/edge/v1/proxy"
	httpxv2 "github.com/sentinez/sentinez/pkg/core/httpx/v2"
	"github.com/valyala/fasthttp"
)

var (
	route     sync.Map
	proxyInst *proxy.Proxy
)

func Store(proxy *proxy.Proxy, config *edgeyaml.Routes) {
	proxyInst = proxy
	for _, routeConfig := range config.Routes {
		// Store the path and target in the sync.Map
		route.Store(routeConfig.Location, routeConfig.ProxyPass)
	}
}

func match(ctx *httpxv2.Context) (string, error) {
	pathRequest := ctx.Path()
	targetRequest := ""

	route.Range(func(location, proxyPass any) bool {
		if strings.HasPrefix(pathRequest, location.(string)) {
			targetRequest = proxyPass.(string)
			// Remove the path prefix from the request path
			remainingPath := strings.TrimPrefix(pathRequest, location.(string))
			ctx.Request.URI().SetPath(remainingPath)

			return false
		}

		return true
	})

	if targetRequest == "" {
		// If no route matches, return an error
		return "", fmt.Errorf("%s", "not found")
	}

	return targetRequest, nil
}

func Match() func(ctx *httpxv2.Context) error {
	return func(ctx *httpxv2.Context) error {
		if proxyInst == nil {
			return ctx.String(fasthttp.StatusInternalServerError,
				"proxy not initialized")
		}

		target, err := match(ctx)
		if err != nil {
			return ctx.String(fasthttp.StatusNotFound, "not found")
		}

		return proxyInst.ServeHTTP(ctx, target)
	}
}
