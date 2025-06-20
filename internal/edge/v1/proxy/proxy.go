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
	"net/http"
	"time"

	httpxv2 "github.com/sentinez/sentinez/pkg/core/httpx/v2"
	"github.com/sentinez/sentinez/pkg/std/zlog"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

type Proxy struct {
	pool proxy.Pool
}

func factory(hostAddr string) (*proxy.ReverseProxy, error) {
	return proxy.NewReverseProxyWith(
		proxy.WithAddress(hostAddr),
		proxy.WithTimeout(10*time.Second),
	)
}

func New() (*Proxy, error) {
	initialCap, maxCap := 100, 1000
	pool, err := proxy.NewChanPool(initialCap, maxCap, factory)
	if err != nil {
		return nil, err
	}

	return &Proxy{pool: pool}, nil
}

func (p *Proxy) ServeHTTP(ctx *httpxv2.Context, target string) error {
	proxyServer, err := p.pool.Get(target)
	if err != nil {
		zlog.Debug("[edge] proxy got an error: ", err)
		ctx.SetStatusCode(http.StatusInternalServerError)
		return err
	}
	defer func() { _ = p.pool.Put(proxyServer) }()
	proxyServer.ServeHTTP(ctx.RequestCtx)

	return nil
}
