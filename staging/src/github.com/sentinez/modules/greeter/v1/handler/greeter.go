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

// Package greeterhdl provides the controller for the greeter service.
package greeterhdl

import (
	"context"
	"fmt"
	"time"

	"github.com/sentinez/shared/zlog"

	greeterpb "github.com/sentinez/sentinez/api/gen/go/sentinez/modules/greeter/v1"
)

// New creates a new Greeter module.
func New() greeterpb.GreeterServiceServer {
	return &Greeter{}
}

// Greeter is the module for Greeter.
type Greeter struct {
	greeterpb.UnimplementedGreeterServiceServer
}

// SayHello implements GreeterServer.
func (g *Greeter) SayHello(_ context.Context,
	msg *greeterpb.SayHelloRequest) (*greeterpb.SayHelloResponse, error) {

	zlog.Debugf("greeter.SayHello: req = %v", msg)

	resp := fmt.Sprintf("Hello %s! Current time is %s",
		msg.Name, time.Now().Format(time.DateTime),
	)

	return &greeterpb.SayHelloResponse{
		Message: resp,
	}, nil
}

// Status healthcheck for consul
func (g *Greeter) Status(ctx context.Context,
	msg *greeterpb.StatusRequest) (*greeterpb.StatusResponse, error) {

	zlog.Debugf("greeter.Status: req = %v", msg)

	_ = ctx
	_ = msg

	return &greeterpb.StatusResponse{
		Message: "ok",
	}, nil
}
