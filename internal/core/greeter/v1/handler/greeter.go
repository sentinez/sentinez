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
	"fmt"
	"time"

	"github.com/sentinez/sentinez/api/gen/go/sentinez/core/greeter/v1"
	"github.com/sentinez/sentinez/pkg/stdcmn/zlog"
)

// New creates a new Greeter module.
func New() greeter.GreeterServiceServer {
	return &Greeter{}
}

// Greeter is the module for Greeter.
type Greeter struct {
	greeter.UnimplementedGreeterServiceServer
}

// SayHello implements GreeterServer.
func (g *Greeter) SayHello(_ context.Context,
	msg *greeter.SayHelloRequest) (*greeter.SayHelloResponse, error) {

	zlog.Debugf("greeter.SayHello: req = %v", msg)

	resp := fmt.Sprintf("Hello %s! Current time is %s",
		msg.Name, time.Now().Format(time.DateTime),
	)

	return &greeter.SayHelloResponse{
		Message: resp,
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
