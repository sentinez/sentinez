// Copyright 2011 The Go Authors.
// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stdproxy

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	corehttp "github.com/sentinez/core/http"
	edgepb "github.com/sentinez/sentinez/api/gen/go/sentinez/edge/v1"
	"github.com/sentinez/sentinez/pkg/network"
	httpxcmn "github.com/sentinez/sentinez/pkg/network/httpx/common"
	stdhttpx "github.com/sentinez/sentinez/pkg/network/httpx/std"
	"github.com/sentinez/shared/zlog"
)

func NewReverseProxy(upstream *edgepb.Upstream) (*ReverseProxy, error) {
	var target string
	switch upstream.GetProtocol() {
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTP:
		target = "http://" + upstream.GetServer()
	case edgepb.ProxyProtocol_PROXY_PROTOCOL_HTTPS:
		target = "https://" + upstream.GetServer()
	default:
		return nil, fmt.Errorf(
			"edge: unsupported protocol: %v", upstream.GetProtocol())
	}

	urlParsed, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	client := httputil.NewSingleHostReverseProxy(urlParsed)
	client.Transport = network.StandardTransporter()

	return &ReverseProxy{
		client: client,
		host:   urlParsed.Host,
	}, nil
}

type ReverseProxy struct {
	client *httputil.ReverseProxy
	host   string
}

func (p *ReverseProxy) Serve(ctx corehttp.Context) {
	if p == nil {
		_ = httpxcmn.NotFound(ctx)
		zlog.Errorf("target not found in reverse proxy memory")
		return
	}

	nctx, ok := ctx.Unwrap().(*stdhttpx.Context)
	if !ok {
		_ = httpxcmn.InternalServerError(ctx)
		zlog.Fatal("request context not supported")
		return
	}

	nctx.SetHost(p.host)

	p.client.ServeHTTP(nctx.Response(), nctx.Request())
}
