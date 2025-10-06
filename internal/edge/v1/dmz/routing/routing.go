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
	"github.com/sentinez/sentinez/internal/edge/v1/cache/routes"
	httpxhz "github.com/sentinez/sentinez/pkg/core/net/httpx/hz"
	"github.com/sentinez/sentinez/pkg/core/net/httpx/hz/proxy"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

func Serve(server httpxhz.Server) error {

	reverseProxy, err := proxy.NewReverseProxy()
	if err != nil {
		zlog.Errorf("failed to create proxy instance: %v", err)
		return err
	}

	handler := routes.GetRouter().SetProxy(reverseProxy)

	server.Handle(handler)

	return nil
}
