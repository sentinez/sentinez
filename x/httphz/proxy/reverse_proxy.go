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

package proxydmz

import (
	"fmt"
	"strings"

	corehttp "github.com/sentinez/sentinez/core/http"
	"github.com/sentinez/sentinez/shared/sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/reverseproxy"
)

var _ corehttp.ReverseProxy = (*ReverseProxy)(nil)

const (
	hzHostClientName = "sentinez-edge-reverse-proxy"
)

func NewReverseProxy(options ...Option) (*ReverseProxy, error) {
	option := defaultBuildOption()
	for _, opt := range options {
		opt.apply(option)
	}

	tlsClient, err := client.NewClient(
		client.WithDialer(standard.NewDialer()),
		client.WithName(hzHostClientName),
		client.WithDisablePathNormalizing(option.disablePathNormalizing),
		client.WithTLSConfig(option.tlsConfig),
		client.WithMaxConnDuration(option.maxConnDuration),
		client.WithResponseBodyStream(option.streamResponseBody),
		client.WithDialTimeout(option.timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("httpxdmz: new reverse proxy failed: %v", err)
	}

	plainClient, err := client.NewClient(
		client.WithDialer(standard.NewDialer()),
		client.WithName(hzHostClientName),
		client.WithDialTimeout(option.timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("httpxdmz: new reverse proxy failed: %v", err)
	}

	proxy := &ReverseProxy{
		tlsClient:   tlsClient,
		plainClient: plainClient,

		rPrxPool: sync.NewPool[reverseproxy.ReverseProxy](),
	}

	return proxy, nil
}

type ReverseProxy struct {
	tlsClient   *client.Client
	plainClient *client.Client

	rPrxPool *sync.Pool[reverseproxy.ReverseProxy]
}

func (p *ReverseProxy) Serve(ctx corehttp.Context, target string) {
	r := p.rPrxPool.Get()
	defer p.rPrxPool.Put(r)

	r.Target = target

	if strings.HasPrefix(target, "https://") {
		r.SetDirector(func(req *protocol.Request) {
			req.SetRequestURI(b2s(JoinURLPath(req, target)))
			req.Header.SetHostBytes(req.URI().Host())
		})
		r.SetClient(p.tlsClient)

	} else {
		r.SetDirector(func(req *protocol.Request) {
			req.SetIsTLS(false)
			req.SetRequestURI(b2s(JoinURLPath(req, target)))
			req.Header.SetHostBytes(req.URI().Host())
		})
		r.SetClient(p.plainClient)
	}

	r.ServeHTTP(ctx.Context(), ctx.Unwrap().(*app.RequestContext))
}
