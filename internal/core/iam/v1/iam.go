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

// Package iamv1 provides the Identity Access Management service.
package iamv1

import (
	"context"

	"github.com/sentinez/sentinez/api/client/local"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/iam/v1"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	iamfac "github.com/sentinez/sentinez/internal/core/iam/v1/factory"
	netgrpc "github.com/sentinez/sentinez/pkg/network/grpc"
	"google.golang.org/grpc/test/bufconn"
)

var bufLis *bufconn.Listener

func GetListener() *bufconn.Listener {
	return bufLis
}

func NewService(ctx context.Context, appConf *confpb.Config) *IAM {
	return &IAM{
		Server: netgrpc.NewDefault(appConf.GetMeta()),
		hdl:    iamfac.NewDefaultHandler(ctx, appConf),
	}
}

// New creates a new Greeter module.

type IAM struct {
	*netgrpc.Server
	hdl iam.IdentityAccessManagementServiceServer
}

func (im *IAM) Start(_ context.Context) error {
	iam.RegisterIdentityAccessManagementServiceServer(im.AsServer(), im.hdl)

	bufLis = bufconn.Listen(local.BufSize)
	return im.BufServe(bufLis)
}
