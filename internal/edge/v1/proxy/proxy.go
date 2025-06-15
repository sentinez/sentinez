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

// Package proxy ...
package proxy

import (
	"log"

	"github.com/sentinez/sentinez/internal/edge/v1/origin"
	httpxv2 "github.com/sentinez/sentinez/pkg/core/httpx/v2"
	"github.com/valyala/fasthttp"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

func Handler(pool proxy.Pool, path string) func(ctx *httpxv2.Context) error {
	return func(ctx *httpxv2.Context) error {
		proxyServer, err := pool.Get(origin.Source(path))
		if err != nil {
			log.Println("ProxyPoolHandler got an error: ", err)
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			return err
		}
		defer func() { _ = pool.Put(proxyServer) }()
		proxyServer.ServeHTTP(ctx.RequestCtx)

		return nil
	}
}

func factory(hostAddr string) (*proxy.ReverseProxy, error) {
	return proxy.NewReverseProxyWith(proxy.WithAddress(hostAddr))
}

func NewChanPool() (proxy.Pool, error) {
	initialCap, maxCap := 100, 1000
	pool, err := proxy.NewChanPool(initialCap, maxCap, factory)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
