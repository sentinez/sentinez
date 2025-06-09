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

// Package greeterhandler provides the controller for the greeter service.
package greeterhandler

import (
	"context"

	domainpb "github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/domain/v1"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	greeterdomain "github.com/sentinez/sentinez/internal/core/greeter/v1/domain"
	"github.com/sentinez/sentinez/pkg/common/copier"
	"github.com/sentinez/sentinez/pkg/std/errors"
	"github.com/sentinez/sentinez/pkg/std/zlog"
)

// New creates a new Greeter module.
func New(biz greeterdomain.IGreeter) greeter.GreeterServiceServer {
	return &Greeter{
		domain: biz,
	}
}

// Greeter is the module for Greeter.
type Greeter struct {
	greeter.UnimplementedGreeterServiceServer
	domain domainpb.GreeterDomainServiceServer
}

// SayHello implements GreeterServer.
func (g *Greeter) SayHello(ctx context.Context,
	msg *greeter.SayHelloRequest) (*greeter.SayHelloResponse, error) {

	var SayHelloReq domainpb.SayHelloRequest
	if err := copier.CopyProtoMessage(msg, &SayHelloReq); err != nil {
		zlog.Error(err)

		return nil, errors.StatusInvalidData
	}

	sayHelloResp, err := g.domain.SayHello(ctx, &SayHelloReq)
	if err != nil {
		return nil, err
	}

	return &greeter.SayHelloResponse{
		Response: sayHelloResp.Response,
	}, nil
}

// Status healthcheck for consul
func (g *Greeter) Status(ctx context.Context,
	msg *greeter.StatusRequest) (*greeter.StatusResponse, error) {

	zlog.Debugf("greeter.Status: req = %v", msg)

	_ = ctx
	_ = msg

	return &greeter.StatusResponse{
		Message: "ok",
	}, nil
}
