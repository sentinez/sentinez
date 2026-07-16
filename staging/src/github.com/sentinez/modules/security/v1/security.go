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

// Package security provides the Security service.
package security

import (
	"context"

	coregrpc "github.com/sentinez/core/grpc"
	securityfac "github.com/sentinez/modules/security/v1/factory"
	"github.com/sentinez/sentinez/api/client/local"
	securitypb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/security/v1"
	settingpb "github.com/sentinez/sentinez/api/gen/go/sentinez/setting/v1"
	"google.golang.org/grpc/test/bufconn"
)

var bufLis *bufconn.Listener

// GetListener returns the local buffer listener for the security module.
func GetListener() *bufconn.Listener {
	return bufLis
}

// NewService creates a new Security module instance.
func NewService(ctx context.Context, appConf *settingpb.Config) *Security {
	return &Security{
		Server: coregrpc.New(coregrpc.WithXMeta(appConf.GetMeta())),
		hdl:    securityfac.NewDefaultHandler(ctx, appConf),
	}
}

// Security orchestrates the security core domain.
type Security struct {
	*coregrpc.Server
	hdl securitypb.SecurityServiceServer
}

// Start registers the handler and begins
// serving gRPC traffic over a buffer listener.
func (mod *Security) Start(_ context.Context) error {
	securitypb.RegisterSecurityServiceServer(mod.AsServer(), mod.hdl)

	// Begin listening on the buffer
	bufLis = bufconn.Listen(local.BufSize)

	// Serve requests over the buffer connections
	return mod.BufServe(bufLis)
}
