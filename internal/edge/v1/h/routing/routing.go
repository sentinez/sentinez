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

	corehttp "github.com/sentinez/core/http"
	"github.com/sentinez/sentinez/pkg/dmz/chains"
)

func NewMockRouter() chains.Handler {
	return &Mock{
		BaseHandler: chains.New(),
	}
}

type Mock struct {
	*chains.BaseHandler
}

func (r *Mock) Handle(ctx corehttp.Context) error {
	return ctx.String(http.StatusOK, "OK")
}

// func NewRouter() chains.Handler {
// 	reverseProxy, err := proxydmz.NewReverseProxy()
// 	if err != nil {
// 		zlog.Errorf("failed to create proxy instance: %v", err)
// 	}

// 	wsReverseProxy, _ := proxydmz.NewWSReverseProxy()

// 	return &Router{
// 		BaseHandler: chains.New(),
// 		httpHandler: routes.GetRouter().SetReverseProxy(reverseProxy),
// 		wsHandler:   routes.GetRouter().SetReverseProxy(wsReverseProxy),
// 	}
// }

// type Router struct {
// 	*chains.BaseHandler
// 	httpHandler func(ctx corehttp.Context) error
// 	wsHandler   func(ctx corehttp.Context) error
// }

// func (r *Router) Handle(ctx corehttp.Context) error {
// 	zlog.Debugf("[edge][%s] >>> visit router", ctx.RequestId())

// 	upgrade := ctx.Header(corehttp.HeaderUpgrade)
// 	if upgrade == "websocket" || upgrade == "WebSocket" {
// 		zlog.Debugf("[edge][websocket] upgrade connection !!!")
// 		return r.wsHandler(ctx)
// 	}

// 	return r.httpHandler(ctx)
// }
